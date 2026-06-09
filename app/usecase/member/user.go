package usecase_member

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	pb "github.com/adkurnwn/gigsourcehub-general-api/proto"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (u *appUsecase) FetchUsers(ctx context.Context, page, limit int64, cursor string, search *string, roleName *string, adminID *string, filter gorm_model.CandidateFilter) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit

	// Calculate totals based on a raw manual query joined explicitly against role system to filter the counter safely before fetching items
	db := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.User{})

	if roleName != nil {
		if *roleName == "Unverified" {
			db = db.Where("users.verified_at IS NULL")
		} else {
			db = db.Joins("JOIN system_roles rs ON users.system_role_id = rs.id").Where("rs.name = ?", *roleName)
			if *roleName == "Candidate" {
				db = db.Where("users.verified_at IS NOT NULL")
			}
		}
	}

	if search != nil && *search != "" {
		db = db.Where("(users.name ILIKE ? OR users.email ILIKE ?)", "%"+*search+"%", "%"+*search+"%")
	}

	if len(filter.Bidang) > 0 {
		db = db.Joins("LEFT JOIN job_titles ON job_titles.id = users.job_title_id").
			Joins("LEFT JOIN sectors ON sectors.id = job_titles.sector_id").
			Where("sectors.name IN ?", filter.Bidang)
	}

	if len(filter.JobRoles) > 0 {
		db = db.Joins("LEFT JOIN user_has_job_roles ON user_has_job_roles.user_id = users.id").
			Joins("LEFT JOIN job_roles ON job_roles.id = user_has_job_roles.job_role_id").
			Where("job_roles.name IN ?", filter.JobRoles)
	}

	if len(filter.CandidateLevel) > 0 {
		db = db.Joins("LEFT JOIN job_titles jt_level ON jt_level.id = users.job_title_id").
			Where("jt_level.name IN ?", filter.CandidateLevel)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count users")
	}

	// Execute actual limited fetch
	var users []gorm_model.User
	if err := db.Preload("SystemRole").
		Preload("RecruitmentStatus").
		Preload("KabupatenKota.Provinsi").
		Preload("JobTitle.Sector").
		Preload("AssignedRole.Sector").
		Preload("JobRoles.Sector").
		Limit(int(limit)).Offset(int(offset)).Order("created_at DESC").Find(&users).Error; err != nil {
		logrus.Error("FetchUsers error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch users")
	}

	// Build bookmark lookup set if adminID is provided
	bookmarkSet := make(map[string]bool)
	if adminID != nil {
		var bookmarkedIDs []string
		u.gormDbRepo.GetDB().WithContext(ctx).
			Model(&gorm_model.Bookmark{}).
			Where("admin_id = ?", *adminID).
			Pluck("candidate_id", &bookmarkedIDs)
		for _, id := range bookmarkedIDs {
			bookmarkSet[id] = true
		}
	}

	var results []interface{}
	for _, user := range users {
		resp := user.ToUserResp()
		if adminID != nil {
			isBookmarked := bookmarkSet[user.ID]
			resp.IsBookmark = &isBookmarked
		}
		results = append(results, resp)
	}

	var nextCursor *string
	if offset+limit < total {
		// As a simple placeholder cursor, we can provide the next page index
		// If the user chooses a more complex cursor later (like a base64 timestamp), they can swap this evaluation block easily.
		nextStr := strconv.FormatInt(page+1, 10)
		nextCursor = &nextStr
	}

	return response.Success(response.List{
		List:   results,
		Limit:  limit,
		Page:   page,
		Total:  total,
		Cursor: nextCursor,
	})
}

func (u *appUsecase) FetchCandidateRecruitment(ctx context.Context, page, limit int64, cursor string, filter gorm_model.CandidateRecruitmentFilter) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit

	query := u.gormDbRepo.GetDB().WithContext(ctx).
		Model(&gorm_model.User{}).
		Joins("JOIN system_roles sr ON sr.id = users.system_role_id").
		Where("sr.name = ? AND users.recruitment_status_id IS NOT NULL", "Candidate")

	if len(filter.JobRoleName) > 0 {
		query = query.Joins("LEFT JOIN user_has_job_roles ON user_has_job_roles.user_id = users.id").
			Joins("LEFT JOIN job_roles ON job_roles.id = user_has_job_roles.job_role_id").
			Where("job_roles.name IN ?", filter.JobRoleName)
	}
	
	if len(filter.CandidateLevel) > 0 {
		query = query.Joins("LEFT JOIN job_titles ON job_titles.id = users.job_title_id").
			Where("job_titles.name IN ?", filter.CandidateLevel)
	}

	if len(filter.ProjectName) > 0 {
		query = query.Joins("LEFT JOIN subrequests ON subrequests.candidate_id = users.id").
			Joins("LEFT JOIN requests ON requests.id = subrequests.request_id").
			Where("requests.project_name IN ?", filter.ProjectName)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count candidates")
	}

	var users []gorm_model.User
	if err := query.
		Preload("SystemRole").
		Preload("RecruitmentStatus").
		Preload("KabupatenKota.Provinsi").
		Preload("JobTitle.Sector").
		Preload("AssignedRole.Sector").
		Preload("JobRoles.Sector").
		Order("users.created_at DESC").
		Limit(int(limit)).Offset(int(offset)).
		Find(&users).Error; err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch candidates")
	}

	results := make([]interface{}, 0, len(users))
	for _, user := range users {
		info, err := u.gormDbRepo.GetActiveSubrequestByCandidateID(ctx, user.ID)
		if err != nil {
			return response.Error(http.StatusInternalServerError, "Failed to fetch candidate subrequest")
		}

		resp := user.ToCandidateTableResp(info)
		results = append(results, resp)
	}

	var nextCursor *string
	if offset+limit < total {
		nextStr := strconv.FormatInt(page+1, 10)
		nextCursor = &nextStr
	}

	return response.Success(response.List{
		List:   results,
		Limit:  limit,
		Page:   page,
		Total:  total,
		Cursor: nextCursor,
	})
}

func (u *appUsecase) FetchCandidateBookmarked(ctx context.Context, adminID string, page, limit int64, cursor string, filter gorm_model.CandidateFilter) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if adminID == "" {
		return response.Error(http.StatusBadRequest, "admin_id is required")
	}

	offset := (page - 1) * limit

	bookmarkQuery := u.gormDbRepo.GetDB().WithContext(ctx).
		Model(&gorm_model.Bookmark{}).
		Where("admin_id = ?", adminID)

	var total int64
	if err := bookmarkQuery.Count(&total).Error; err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count bookmarked candidates")
	}

	var bookmarks []gorm_model.Bookmark
	if err := bookmarkQuery.
		Order("created_at DESC").
		Limit(int(limit)).Offset(int(offset)).
		Find(&bookmarks).Error; err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch bookmarked candidates")
	}

	bookmarkedIDs := make([]string, 0, len(bookmarks))
	for _, bookmark := range bookmarks {
		bookmarkedIDs = append(bookmarkedIDs, bookmark.CandidateID)
	}

	usersByID := make(map[string]gorm_model.User)
	if len(bookmarkedIDs) > 0 {
		userQuery := u.gormDbRepo.GetDB().WithContext(ctx).
			Model(&gorm_model.User{}).
			Where("users.id IN ?", bookmarkedIDs)

		if len(filter.Bidang) > 0 {
			userQuery = userQuery.Joins("LEFT JOIN job_titles ON job_titles.id = users.job_title_id").
				Joins("LEFT JOIN sectors ON sectors.id = job_titles.sector_id").
				Where("sectors.name IN ?", filter.Bidang)
		}

		if len(filter.JobRoles) > 0 {
			userQuery = userQuery.Joins("LEFT JOIN user_has_job_roles ON user_has_job_roles.user_id = users.id").
				Joins("LEFT JOIN job_roles ON job_roles.id = user_has_job_roles.job_role_id").
				Where("job_roles.name IN ?", filter.JobRoles)
		}

		if len(filter.CandidateLevel) > 0 {
			userQuery = userQuery.Joins("LEFT JOIN job_titles jt_level ON jt_level.id = users.job_title_id").
				Where("jt_level.name IN ?", filter.CandidateLevel)
		}

		var users []gorm_model.User
		if err := userQuery.
			Preload("SystemRole").
			Preload("RecruitmentStatus").
			Preload("KabupatenKota.Provinsi").
			Preload("JobTitle.Sector").
			Preload("AssignedRole.Sector").
			Preload("JobRoles.Sector").
			Find(&users).Error; err != nil {
			return response.Error(http.StatusInternalServerError, "Failed to fetch bookmarked candidate data")
		}

		for _, user := range users {
			usersByID[user.ID] = user
		}
	}

	results := make([]interface{}, 0, len(bookmarkedIDs))
	for _, candidateID := range bookmarkedIDs {
		user, ok := usersByID[candidateID]
		if !ok {
			continue
		}

		info, err := u.gormDbRepo.GetActiveSubrequestByCandidateID(ctx, candidateID)
		if err != nil {
			return response.Error(http.StatusInternalServerError, "Failed to fetch candidate subrequest")
		}

		resp := user.ToCandidateTableResp(info)
		isBookmark := true
		resp.IsBookmark = &isBookmark
		results = append(results, resp)
	}

	var nextCursor *string
	if offset+limit < total {
		nextStr := strconv.FormatInt(page+1, 10)
		nextCursor = &nextStr
	}

	return response.Success(response.List{
		List:   results,
		Limit:  limit,
		Page:   page,
		Total:  total,
		Cursor: nextCursor,
	})
}

func (u *appUsecase) FetchUserDetail(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})

	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}

	// Pre-hydrate role system safely
	roleName, errRole := u.gormDbRepo.GetRoleNameByUserID(ctx, user.ID)
	if errRole == nil && roleName != "" {
		user.SystemRole = &gorm_model.SystemRole{
			Name: roleName,
		}
	}

	res := gorm_model.UserDetailResp{
		UserResp: user.ToUserResp(),
	}

	// Fetch CV mapped to user
	cv, errCv := u.gormDbRepo.GetCVByUserID(ctx, user.ID)
	if errCv == nil && cv != nil {
		// Generate 1-hour presigned view link
		expireDuration := time.Hour
		presignedLink := u.storageRepo.GetPresignedLink(cv.Path, &expireDuration)

		res.CV = &gorm_model.CVPrivateResp{
			ID:        cv.ID,
			Name:      cv.Filename,
			URL:       presignedLink,
			CreatedAt: cv.CreatedAt,
		}
	}

	return response.Success(res)
}

func (u *appUsecase) GetProfile(ctx context.Context, claim domain.JWTClaimUser) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: claim.UserID},
	})

	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}

	// Pre-hydrate role system safely
	roleName, errRole := u.gormDbRepo.GetRoleNameByUserID(ctx, user.ID)
	if errRole == nil && roleName != "" {
		user.SystemRole = &gorm_model.SystemRole{
			Name: roleName,
		}
	}

	return response.Success(user.ToUserResp())
}

func (u *appUsecase) CreateBySuperadmin(ctx context.Context, req request_model.CreateUserBySuperadminRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if req.Email == "" || !helpers.IsValidEmail(req.Email) {
		return response.Error(http.StatusBadRequest, "invalid email format")
	}

	// check the db
	existingUser, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		Email: &req.Email,
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if existingUser != nil {
		return response.Error(http.StatusBadRequest, "email already taken")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to hash password")
	}

	if req.JobTitleId != nil && *req.JobTitleId != "" {
		var jt gorm_model.JobTitle
		if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&jt, "id = ?", *req.JobTitleId).Error; err != nil {
			return response.Error(http.StatusBadRequest, "Job title not found")
		}
		if !jt.IsActive {
			return response.Error(http.StatusBadRequest, "Cannot reference an inactive job title")
		}
	}
	if req.AssignedRoleId != nil && *req.AssignedRoleId != "" {
		var jr gorm_model.JobRole
		if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&jr, "id = ?", *req.AssignedRoleId).Error; err != nil {
			return response.Error(http.StatusBadRequest, "Assigned job role not found")
		}
		if !jr.IsActive {
			return response.Error(http.StatusBadRequest, "Cannot reference an inactive job role")
		}
	}

	// Strict Validation for HR Restriction
	if req.SystemRoleId != nil && req.JobTitleId != nil {
		var role gorm_model.SystemRole
		if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&role, "id = ?", *req.SystemRoleId).Error; err == nil {
			var jt gorm_model.JobTitle
			if err := u.gormDbRepo.GetDB().WithContext(ctx).Preload("Sector").First(&jt, "id = ?", *req.JobTitleId).Error; err == nil {
				if role.Name == "Admin" {
					// Admin MUST be in Human Resources sector
					if jt.Sector == nil || jt.Sector.Name != "Human Resources" {
						return response.Error(http.StatusBadRequest, "Admin role must be in Human Resources sector")
					}
				} else if role.Name == "Employee" {
					// Employee (Pegawai) MUST NOT be in Human Resources sector
					if jt.Sector != nil && jt.Sector.Name == "Human Resources" {
						return response.Error(http.StatusBadRequest, "Employee role cannot be in Human Resources sector")
					}
				}
			}
		}
	}

	now := time.Now()
	user := gorm_model.User{
		ID:                uuid.New().String(),
		Email:             req.Email,
		Name:              req.Name,
		Password:          string(hashedPassword),
		AssignedRoleId:    req.AssignedRoleId,
		SystemRoleId:      req.SystemRoleId,
		JobTitleId:        req.JobTitleId,
		VerifiedAt:        &now,
		MustResetPassword: true,
	}

	if req.AccountStatus != nil {
		user.AccountStatus = req.AccountStatus
	}

	if err := u.gormDbRepo.CreateUser(ctx, &user); err != nil {
		helpers.LogActivity(ctx, u.gormDbRepo, "Create", "User Management", user.Email, req, false)
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Create", "User Management", user.Email, req, true)
	return response.Success(user.ToUserResp())
}

func (u *appUsecase) EditUserBySuperadmin(ctx context.Context, id string, req request_model.EditUserBySuperadminRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}

	if req.JobTitleId != nil && *req.JobTitleId != "" {
		if user.JobTitleId == nil || *user.JobTitleId != *req.JobTitleId {
			var jt gorm_model.JobTitle
			if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&jt, "id = ?", *req.JobTitleId).Error; err != nil {
				return response.Error(http.StatusBadRequest, "Job title not found")
			}
			if !jt.IsActive {
				return response.Error(http.StatusBadRequest, "Cannot reference an inactive job title")
			}
		}
	}
	if req.AssignedRoleId != nil && *req.AssignedRoleId != "" {
		if user.AssignedRoleId == nil || *user.AssignedRoleId != *req.AssignedRoleId {
			var jr gorm_model.JobRole
			if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&jr, "id = ?", *req.AssignedRoleId).Error; err != nil {
				return response.Error(http.StatusBadRequest, "Assigned job role not found")
			}
			if !jr.IsActive {
				return response.Error(http.StatusBadRequest, "Cannot reference an inactive job role")
			}
		}
	}

	// Strict Validation for HR Restriction during Edit
	checkRoleId := user.SystemRoleId
	if req.SystemRoleId != nil {
		checkRoleId = req.SystemRoleId
	}
	checkJobTitleId := user.JobTitleId
	if req.JobTitleId != nil {
		checkJobTitleId = req.JobTitleId
	}

	if checkRoleId != nil && checkJobTitleId != nil {
		var role gorm_model.SystemRole
		if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&role, "id = ?", *checkRoleId).Error; err == nil {
			var jt gorm_model.JobTitle
			if err := u.gormDbRepo.GetDB().WithContext(ctx).Preload("Sector").First(&jt, "id = ?", *checkJobTitleId).Error; err == nil {
				if role.Name == "Admin" {
					if jt.Sector == nil || jt.Sector.Name != "Human Resources" {
						return response.Error(http.StatusBadRequest, "Admin role must be in Human Resources sector")
					}
				} else if role.Name == "Employee" {
					if jt.Sector != nil && jt.Sector.Name == "Human Resources" {
						return response.Error(http.StatusBadRequest, "Employee role cannot be in Human Resources sector")
					}
				}
			}
		}
	}

	if req.Name != nil && *req.Name != "" {
		user.Name = *req.Name
	}

	if req.AssignedRoleId != nil {
		user.AssignedRoleId = req.AssignedRoleId
	}

	if req.SystemRoleId != nil {
		user.SystemRoleId = req.SystemRoleId
	}

	if req.JobTitleId != nil {
		user.JobTitleId = req.JobTitleId
	}

	if req.AccountStatus != nil {
		user.AccountStatus = req.AccountStatus
	}

	if err := u.gormDbRepo.UpdateUser(ctx, user); err != nil {
		helpers.LogActivity(ctx, u.gormDbRepo, "Edit", "User Management", user.Email, req, false)
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Edit", "User Management", user.Email, req, true)
	return response.Success(user.ToUserResp())
}

func (u *appUsecase) BlockUserBySuperadmin(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}
	if *user.AccountStatus == "Blocked" {
		return response.Error(http.StatusBadRequest, "User already blocked")
	}

	if user.SystemRole.Name == "Admin" || user.SystemRole.Name == "Employee" || user.SystemRole.Name == "Superadmin" {
		return response.Error(http.StatusBadRequest, "Admin, Employee, and Superadmin users cannot be blocked")
	}

	blockedStatus := "Blocked"
	user.AccountStatus = &blockedStatus

	if err := u.gormDbRepo.UpdateUser(ctx, user); err != nil {
		helpers.LogActivity(ctx, u.gormDbRepo, "Block", "User Management", user.Email, nil, false)
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Block", "User Management", user.Email, nil, true)
	return response.SuccessAction("User", user.Email, "blocked")
}

func (u *appUsecase) DisableUserBySuperadmin(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}
	if *user.AccountStatus == "Inactive" {
		return response.Error(http.StatusBadRequest, "User already inactive")
	}

	if user.SystemRole.Name == "Candidate" {
		return response.Error(http.StatusBadRequest, "Candidate user cannot be disabled")
	}

	inactiveStatus := "Inactive"
	user.AccountStatus = &inactiveStatus

	if err := u.gormDbRepo.UpdateUser(ctx, user); err != nil {
		helpers.LogActivity(ctx, u.gormDbRepo, "Disable", "User Management", user.Email, nil, false)
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Disable", "User Management", user.Email, nil, true)
	return response.SuccessAction("User", user.Email, "disabled")
}

func (u *appUsecase) ActivateUserBySuperadmin(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}

	oldStatus := user.AccountStatus
	activeStatus := "Active"
	user.AccountStatus = &activeStatus

	if err := u.gormDbRepo.UpdateUser(ctx, user); err != nil {
		helpers.LogActivity(ctx, u.gormDbRepo, "Activate", "User Management", user.Email, nil, false)
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Activate", "User Management", user.Email, nil, true)
	if *oldStatus == "Blocked" {
		return response.SuccessAction("User", user.Email, "ublocked")
	}

	return response.SuccessAction("User", user.Email, "activated")
}

func (u *appUsecase) UpdateProfile(ctx context.Context, userID string, req request_model.UpdateProfileRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: userID},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}

	// Update fields
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Birthdate != nil {
		t, err := time.Parse("2006-01-02", *req.Birthdate)
		if err == nil {
			user.Birthdate = &t
		}
	}
	if req.SchoolUniversity != nil {
		user.SchoolUniversity = req.SchoolUniversity
	}
	if req.Major != nil {
		user.Major = req.Major
	}
	if req.Gpa != nil {
		user.Gpa = req.Gpa
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = req.PhoneNumber
	}
	if req.PortofolioLink != nil {
		user.PortofolioLink = req.PortofolioLink
	}
	if req.KabupatenKotaId != nil {
		user.KabupatenKotaId = req.KabupatenKotaId
	}
	if req.YearsExperience != nil {
		user.YearsExperience = req.YearsExperience
	}
	if req.TechStack != nil {
		techStackJson, _ := json.Marshal(req.TechStack)
		tsStr := string(techStackJson)
		user.TechStack = &tsStr
	}

	if req.Summary != nil {
		user.Summary = req.Summary
	}

	if req.AvailabilityStatus != nil {
		isAllowed := true
		if user.RecruitmentStatusId != nil && user.RecruitmentStatus != nil {
			if user.RecruitmentStatus.Name != "Available" && user.RecruitmentStatus.Name != "Unavailable" {
				isAllowed = false
			}
		}
		if !isAllowed {
			return response.Error(http.StatusBadRequest, "Anda tidak dapat mengubah status ketersediaan saat sedang dalam proses rekrutmen atau onboarding")
		}

		if *req.AvailabilityStatus == "available" {
			user.UnavailableUntil = nil
			user.RecruitmentStatusId = nil
		} else if *req.AvailabilityStatus == "unavailable" && req.UnavailableUntil != nil {
			if *req.UnavailableUntil == "" {
				user.UnavailableUntil = nil
				user.RecruitmentStatusId = nil
			} else {
				t, err := time.Parse("2006-01-02", *req.UnavailableUntil)
				if err == nil {
					user.UnavailableUntil = &t
					var unavailableStatus gorm_model.RecruitmentStatus
					if err := u.gormDbRepo.GetDB().WithContext(ctx).
						Where("name = ? AND deleted_at IS NULL", "Unavailable").
						First(&unavailableStatus).Error; err == nil {
						user.RecruitmentStatusId = &unavailableStatus.ID
					} else {
						return response.Error(http.StatusInternalServerError, "Gagal mendapatkan data status recruitment 'Unavailable'")
					}
				} else {
					return response.Error(http.StatusBadRequest, "Format tanggal tidak valid. Gunakan format YYYY-MM-DD")
				}
			}
		}
	}

	if req.JobRoleIds != nil {
		var jobRoles []gorm_model.JobRole
		if len(req.JobRoleIds) > 0 {
			db := u.gormDbRepo.GetDB()

			var allRoles []gorm_model.JobRole
			db.Preload("Sector").Find(&allRoles)

			var dbRoleUUIDs []string
			var newRoleStrings []string

			for _, idOrNew := range req.JobRoleIds {
				if strings.HasPrefix(idOrNew, "NEW_ROLE:") {
					newRoleStrings = append(newRoleStrings, idOrNew)
				} else if _, err := uuid.Parse(idOrNew); err == nil {
					dbRoleUUIDs = append(dbRoleUUIDs, idOrNew)
				}
			}

			if len(dbRoleUUIDs) > 0 {
				var existingRoles []gorm_model.JobRole
				if err := db.WithContext(ctx).Where("id IN ?", dbRoleUUIDs).Find(&existingRoles).Error; err != nil {
					logrus.Errorf("failed to fetch job roles: %v", err)
				} else {
					jobRoles = append(jobRoles, existingRoles...)
				}
			}

			for _, newRoleStr := range newRoleStrings {
				parts := strings.Split(newRoleStr, ":")
				var sectorName string
				var roleName string
				if len(parts) >= 3 {
					sectorName = parts[1]
					roleName = parts[2]
				} else if len(parts) >= 2 {
					roleName = parts[1]
				}

				if roleName == "" {
					continue
				}

				if match, found := fuzzyMatchJobRole(allRoles, roleName, 2); found {
					alreadyAdded := false
					for _, jr := range jobRoles {
						if jr.ID == match.ID {
							alreadyAdded = true
							break
						}
					}
					if !alreadyAdded {
						jobRoles = append(jobRoles, *match)
					}
					continue
				}

				titled := toTitleCase(roleName)
				var sector *gorm_model.Sector
				var err error

				if sectorName != "" {
					sector, err = getOrCreateSectorByName(db, sectorName)
				} else {
					sector, err = getOrCreateUndefinedSector(db)
				}

				if err == nil {
					newRole := gorm_model.JobRole{
						ID:       uuid.New().String(),
						Name:     titled,
						SectorID: sector.ID,
					}
					if err := db.Create(&newRole).Error; err == nil {
						jobRoles = append(jobRoles, newRole)
						allRoles = append(allRoles, newRole)
					}
				}
			}

			existingRoleMap := make(map[string]bool)
			for _, r := range user.JobRoles {
				existingRoleMap[r.ID] = true
			}
			for _, jr := range jobRoles {
				if !jr.IsActive && !existingRoleMap[jr.ID] {
					return response.Error(http.StatusBadRequest, "Cannot reference inactive job role: " + jr.Name)
				}
			}
		}

		if err := u.gormDbRepo.GetDB().WithContext(ctx).Model(user).Association("JobRoles").Replace(jobRoles); err != nil {
			logrus.Errorf("failed to update job roles association: %v", err)
		}

		user.JobRoles = jobRoles
	}

	if err := u.gormDbRepo.UpdateUser(ctx, user); err != nil {
		helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Profile", user.Email, req, false)
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Profile", user.Email, req, true)

	// Synchronize with Qdrant
	go func() {
		syncCtx := context.Background()

		location := ""
		if user.KabupatenKotaId != nil {
			kab, _ := u.gormDbRepo.GetKabupatenName(syncCtx, *user.KabupatenKotaId)
			provId := ""
			u.gormDbRepo.GetDB().Table("kabupaten_kota").Where("id = ?", *user.KabupatenKotaId).Pluck("provinsi_id", &provId)
			prov, _ := u.gormDbRepo.GetProvinsiName(syncCtx, provId)
			if kab != "" && prov != "" {
				location = fmt.Sprintf("%s, %s", kab, prov)
			} else if kab != "" {
				location = kab
			}
		}

		summary := ""
		if user.Summary != nil && *user.Summary != "" {
			summary = *user.Summary
		} else {
			cv, err := u.gormDbRepo.GetCVByUserID(syncCtx, user.ID)
			if err == nil && cv != nil && cv.ParsedData != nil {
				var parsed map[string]interface{}
				if err := json.Unmarshal([]byte(*cv.ParsedData), &parsed); err == nil {
					if s, ok := parsed["summary"].(string); ok {
						summary = s
					}
				}
			}
		}

		var roleNames []string
		u.gormDbRepo.GetDB().Table("job_roles").
			Joins("JOIN user_has_job_roles ON job_roles.id = user_has_job_roles.job_role_id").
			Where("user_has_job_roles.user_id = ?", user.ID).
			Pluck("name", &roleNames)

		var techStack []string
		if user.TechStack != nil {
			json.Unmarshal([]byte(*user.TechStack), &techStack)
		}

		gpa := 0.0
		if user.Gpa != nil {
			gpa = *user.Gpa
		}

		major := ""
		if user.Major != nil {
			major = *user.Major
		}

		yearsExp := 0
		if user.YearsExperience != nil {
			yearsExp = *user.YearsExperience
		}

		updateReq := &pb.UpdateCandidateRequest{
			UserId:          user.ID,
			Name:            user.Name,
			Roles:           roleNames,
			Summary:         summary,
			EducationMajor:  major,
			Gpa:             gpa,
			Location:        location,
			TechStack:       techStack,
			YearsExperience: int32(yearsExp),
		}

		if u.aiSearchRepo != nil {
			if err := u.aiSearchRepo.UpdateCandidate(syncCtx, updateReq); err != nil {
				logrus.Errorf("failed to synchronize candidate update to AI API: %v", err)
			}
		}
	}()

	return response.Success(user.ToUserResp())
}

func (u *appUsecase) FetchUserThumb(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})

	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}

	if user.ProfilePicture == nil || *user.ProfilePicture == "" {
		return response.Success(map[string]interface{}{
			"profile_picture_url": nil,
		})
	}

	pp := *user.ProfilePicture
	// Construct full public URL if not already a full URL
	if len(pp) > 0 && pp[0] != 'h' {
		pp = u.storageRepo.GetPublicLink(pp)
	}

	// Apply _thumb suffix logic before extension
	thumbURL := pp
	lastDotIndex := -1
	for i := len(pp) - 1; i >= 0; i-- {
		if pp[i] == '.' {
			lastDotIndex = i
			break
		}
		if pp[i] == '/' {
			break
		}
	}

	if lastDotIndex != -1 {
		filename := pp[:lastDotIndex]
		extension := pp[lastDotIndex:]
		thumbURL = fmt.Sprintf("%s_thumb%s", filename, extension)
	}

	return response.Success(map[string]interface{}{
		"profile_picture_url": thumbURL,
	})
}

func (u *appUsecase) UpdatePassword(ctx context.Context, userID string, req request_model.UpdatePasswordRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: userID},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return response.Error(http.StatusBadRequest, "Old password does not match")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to hash new password")
	}

	user.Password = string(hashedPassword)
	user.MustResetPassword = false

	if err := u.gormDbRepo.UpdateUser(ctx, user); err != nil {
		helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Password", user.Email, nil, false)
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Password", user.Email, nil, true)
	return response.SuccessAction("User", user.Email, "password updated")
}

func (u *appUsecase) PatchUserRecruitmentStatus(ctx context.Context, id string, req request_model.PatchUserRecruitmentStatusRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}

	// Validate recruitment status ID if provided and changed
	if req.RecruitmentStatusId != nil && *req.RecruitmentStatusId != "" {
		if user.RecruitmentStatusId == nil || *user.RecruitmentStatusId != *req.RecruitmentStatusId {
			var recruitmentStatus gorm_model.RecruitmentStatus
			if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&recruitmentStatus, "id = ?", *req.RecruitmentStatusId).Error; err != nil {
				return response.Error(http.StatusBadRequest, "Invalid recruitment status ID")
			}
			if !recruitmentStatus.IsActive {
				return response.Error(http.StatusBadRequest, "Cannot reference an inactive recruitment status")
			}
		}
	}

	// Update candidate level if provided
	if req.CandidateLevel != nil {
		user.CandidateLevel = req.CandidateLevel
	}

	// Update recruitment status ID if provided
	if req.RecruitmentStatusId != nil {
		user.RecruitmentStatusId = req.RecruitmentStatusId

		changerID := helpers.GetActorID(ctx)
		if changerID != "" {
			var subReqID *string
			if activeSR, errSR := u.gormDbRepo.GetActiveSubrequestByCandidateID(ctx, user.ID); errSR == nil && activeSR != nil {
				subReqID = &activeSR.SubrequestID
			}
			var statusID *string
			if *req.RecruitmentStatusId != "" {
				statusID = req.RecruitmentStatusId
			}
			history := &gorm_model.CandidateStatusHistory{
				CandidateUserID:     user.ID,
				RecruitmentStatusID: statusID,
				SubrequestID:        subReqID,
				ChangedByUserID:     changerID,
			}
			if errHist := u.gormDbRepo.CreateCandidateStatusHistory(ctx, history); errHist != nil {
				logrus.Errorf("failed to create status history: %v", errHist)
			}
		}
	}

	if err := u.gormDbRepo.UpdateUser(ctx, user); err != nil {
		helpers.LogActivity(ctx, u.gormDbRepo, "Patch", "User Recruitment Status", user.Email, req, false)
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Patch", "User Recruitment Status", user.Email, req, true)
	return response.Success(user.ToUserResp())
}

func (u *appUsecase) getRecruitmentStatusByName(ctx context.Context, statusName string) (*gorm_model.RecruitmentStatus, error) {
	var recruitmentStatus gorm_model.RecruitmentStatus
	if err := u.gormDbRepo.GetDB().WithContext(ctx).
		Where("name = ? AND deleted_at IS NULL", statusName).
		First(&recruitmentStatus).Error; err != nil {
		return nil, err
	}

	return &recruitmentStatus, nil
}

func (u *appUsecase) DeclineRecruitment(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}
	if user.SystemRole == nil || user.SystemRole.Name != "Candidate" {
		return response.Error(http.StatusBadRequest, "User is not a candidate")
	}

	declineStatus, err := u.getRecruitmentStatusByName(ctx, "Decline")
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Decline recruitment status not found")
	}
	if !declineStatus.IsActive {
		return response.Error(http.StatusBadRequest, "Cannot reference an inactive recruitment status")
	}

	user.RecruitmentStatusId = &declineStatus.ID
	if err := u.gormDbRepo.UpdateUser(ctx, user); err != nil {
		helpers.LogActivity(ctx, u.gormDbRepo, "Patch", "User Recruitment Status", user.Email, nil, false)
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Patch", "User Recruitment Status", user.Email, nil, true)
	
	// Notify Admin
	go func() {
		bgCtx := context.Background()
		subReq, err := u.gormDbRepo.GetActiveSubrequestByCandidateID(bgCtx, user.ID)
		if err == nil && subReq != nil {
			var req gorm_model.Request
			if errReq := u.gormDbRepo.GetDB().WithContext(bgCtx).First(&req, "id = ?", subReq.RequestID).Error; errReq == nil && req.AdminUserID != nil {
				title := "Kandidat Menolak Rekrutmen"
				desc := fmt.Sprintf("Kandidat %s menolak rekrutmen.", user.Name)
				helpers.SendNotificationAsync(bgCtx, u.gormDbRepo, *req.AdminUserID, title, desc)
			}
		}
	}()
	
	return response.SuccessAction("User", user.Email, "recruitment declined")
}

func (u *appUsecase) ConfirmDeclineRecruitment(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}
	if user.SystemRole == nil || user.SystemRole.Name != "Candidate" {
		return response.Error(http.StatusBadRequest, "User is not a candidate")
	}

	statusName, err := u.gormDbRepo.GetCandidateRecruitmentStatusName(ctx, user.ID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch candidate recruitment status")
	}
	if statusName != "Decline" {
		return response.Error(http.StatusBadRequest, "User recruitment status must be Decline")
	}

	if err := u.gormDbRepo.CancelRecruitmentByCandidateID(ctx, user.ID); err != nil {
		logrus.Errorf("ConfirmDeclineRecruitment failed for user %s: %v", user.ID, err)
		return response.Error(http.StatusInternalServerError, "Failed to confirm decline")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Patch", "User Recruitment Status", user.Email, nil, true)
	return response.SuccessAction("User", user.Email, "decline confirmed")
}

func (u *appUsecase) StopOnboarding(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}
	if user.SystemRole == nil || user.SystemRole.Name != "Candidate" {
		return response.Error(http.StatusBadRequest, "User is not a candidate")
	}

	statusName, err := u.gormDbRepo.GetCandidateRecruitmentStatusName(ctx, user.ID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch candidate recruitment status")
	}
	if statusName != "Accepted" {
		return response.Error(http.StatusBadRequest, "User recruitment status must be Accepted")
	}

	if err := u.gormDbRepo.StopOnboardingByCandidateID(ctx, user.ID); err != nil {
		logrus.Errorf("StopOnboarding failed for user %s: %v", user.ID, err)
		return response.Error(http.StatusInternalServerError, "Failed to stop onboarding")
	}

	go func(email, name string) {
		if err := u.mailerRepo.SendStopOnboardingEmail(email, name); err != nil {
			logrus.Errorf("Failed to send stop onboarding email to %s: %v", email, err)
		}
	}(user.Email, user.Name)

	helpers.LogActivity(ctx, u.gormDbRepo, "Patch", "User Recruitment Status", user.Email, nil, true)
	return response.SuccessAction("User", user.Email, "onboarding stopped")
}

func (u *appUsecase) CancelRecruitment(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}
	if user.SystemRole == nil || user.SystemRole.Name != "Candidate" {
		return response.Error(http.StatusBadRequest, "User is not a candidate")
	}

	if err := u.gormDbRepo.CancelRecruitmentByCandidateID(ctx, user.ID); err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to cancel recruitment")
	}

	// Send Email asynchronously
	go func(email, name string) {
		if err := u.mailerRepo.SendCancelRecruitmentEmail(email, name); err != nil {
			logrus.Errorf("Failed to send cancel recruitment email to %s: %v", email, err)
		}
	}(user.Email, user.Name)

	// Schedule chat deletion in 12 hours
	candidateID := user.ID
	time.AfterFunc(12*time.Hour, func() {
		bgCtx := context.Background()
		if err := u.gormDbRepo.DeleteConversationsByCandidateID(bgCtx, candidateID); err != nil {
			logrus.Errorf("Failed to delete conversations for candidate %s after 12 hours: %v", candidateID, err)
		} else {
			logrus.Infof("Successfully deleted conversations for candidate %s after 12 hours", candidateID)
		}
	})

	// Notify Candidate
	helpers.SendNotificationAsync(ctx, u.gormDbRepo, user.ID,
		"Recruitment Status Update",
		"We regret to inform you that you did not pass this recruitment stage.",
	)

	return response.SuccessAction("User", user.Email, "recruitment canceled")
}

func (u *appUsecase) FinalizeRecruitment(ctx context.Context, adminID string, req request_model.FinalizeRecruitmentRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if req.CandidateUserID == "" || req.SubrequestID == "" {
		return response.Error(http.StatusBadRequest, "candidate_user_id and subrequest_id are required")
	}
	if req.StartDate == nil || *req.StartDate == "" {
		return response.Error(http.StatusBadRequest, "start_date is required")
	}
	if req.EndDate == nil || *req.EndDate == "" {
		return response.Error(http.StatusBadRequest, "end_date is required")
	}

	startDate, err := time.Parse("2006-01-02", *req.StartDate)
	if err != nil {
		return response.Error(http.StatusBadRequest, "invalid start_date format")
	}
	endDate, err := time.Parse("2006-01-02", *req.EndDate)
	if err != nil {
		return response.Error(http.StatusBadRequest, "invalid end_date format")
	}
	if endDate.Before(startDate) {
		return response.Error(http.StatusBadRequest, "end_date must be after start_date")
	}

	userRole, err := u.gormDbRepo.GetRoleNameByUserID(ctx, adminID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to verify requester role")
	}
	if userRole != "Admin" {
		return response.Error(http.StatusForbidden, "Only Admin can finalize recruitment")
	}

	adminAllowed, err := u.gormDbRepo.IsAdminOfSubrequest(ctx, adminID, req.SubrequestID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to verify admin assignment")
	}
	if !adminAllowed {
		return response.Error(http.StatusForbidden, "You are not assigned to this subrequest's request")
	}

	candidate, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: req.CandidateUserID},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch candidate")
	}
	if candidate == nil {
		return response.Error(http.StatusNotFound, "Candidate not found")
	}
	if candidate.SystemRole == nil || candidate.SystemRole.Name != "Candidate" {
		return response.Error(http.StatusBadRequest, "User is not a candidate")
	}

	isCandidate, err := u.gormDbRepo.IsCandidateOnSubrequest(ctx, req.CandidateUserID, req.SubrequestID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to verify candidate assignment")
	}
	if !isCandidate {
		return response.Error(http.StatusBadRequest, "Candidate is not assigned to this subrequest")
	}

	snapshotData, err := u.gormDbRepo.GetFinalizeSnapshotData(ctx, req.SubrequestID)
	if err != nil {
		if err.Error() == "record not found" {
			return response.Error(http.StatusNotFound, "Subrequest snapshot data not found")
		}
		logrus.Errorf("GetFinalizeSnapshotData error: %v", err)
		return response.Error(http.StatusInternalServerError, "Failed to build snapshot data")
	}

	snapshotPayload := map[string]interface{}{
		"subrequest_id":    snapshotData.SubrequestID,
		"job_role_id":      snapshotData.JobRoleID,
		"job_role_name":    snapshotData.JobRoleName,
		"project_name":     snapshotData.ProjectName,
		"employee_user_id": snapshotData.EmployeeUserID,
		"employee_name":    snapshotData.EmployeeName,
	}
	encodedSnapshot, err := json.Marshal(snapshotPayload)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to serialize snapshot")
	}

	acceptedStatusID := "1db9ec40-fdc0-4357-9d7a-1ed6f38fe1cb"
	var acceptedStatus gorm_model.RecruitmentStatus
	if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&acceptedStatus, "id = ?", acceptedStatusID).Error; err != nil {
		return response.Error(http.StatusInternalServerError, "Accepted recruitment status not found")
	}

	if err := u.gormDbRepo.FinalizeRecruitment(
		ctx,
		req.CandidateUserID,
		req.SubrequestID,
		snapshotData.RequestID,
		acceptedStatus.ID,
		&startDate,
		&endDate,
		nil,
		string(encodedSnapshot),
	); err != nil {
		helpers.LogActivity(ctx, u.gormDbRepo, "Finalize", "Recruitment", candidate.Email, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to finalize recruitment")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Finalize", "Recruitment", candidate.Email, req, true)
	return response.SuccessAction("User", candidate.Email, "recruitment finalized")
}

func (u *appUsecase) GetActiveSubrequest(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}
	if user.SystemRole == nil || user.SystemRole.Name != "Candidate" {
		return response.Error(http.StatusBadRequest, "User is not a candidate")
	}

	info, err := u.gormDbRepo.GetActiveSubrequestByCandidateID(ctx, user.ID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch active subrequest")
	}
	if info == nil {
		return response.Error(http.StatusNotFound, "Active subrequest not found")
	}

	return response.Success(info)
}

func (u *appUsecase) DeleteAccount(ctx context.Context, userID string, req request_model.DeleteAccountRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if req.Password == "" {
		return response.Error(http.StatusBadRequest, "Password is required")
	}

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: userID},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}

	// Only Candidate can self-delete
	roleName, errRole := u.gormDbRepo.GetRoleNameByUserID(ctx, user.ID)
	if errRole != nil {
		return response.Error(http.StatusInternalServerError, "Failed to verify user role")
	}
	if roleName != "Candidate" {
		return response.Error(http.StatusForbidden, "Only candidates can delete their own account")
	}

	// Verify password confirmation
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return response.Error(http.StatusBadRequest, "Password is incorrect")
	}

	// Soft-delete the user (sets deleted_at via GORM)
	if err := u.gormDbRepo.SoftDeleteUser(ctx, user.ID); err != nil {
		helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "Account", user.Email, nil, false)
		return response.Error(http.StatusInternalServerError, "Failed to delete account")
	}

	// Synchronise deletion with Qdrant
	go func() {
		syncCtx := context.Background()
		if u.aiSearchRepo != nil {
			if err := u.aiSearchRepo.DeleteCandidate(syncCtx, user.ID); err != nil {
				logrus.Errorf("failed to synchronize candidate deletion to AI API: %v", err)
			}
		}
	}()

	// Revoke all refresh tokens so new access tokens cannot be generated
	_ = u.gormDbRepo.DeleteUserTokensByUserID(ctx, user.ID, gorm_model.TokenTypeRefresh)

	helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "Account", user.Email, nil, true)
	return response.SuccessAction("Account", user.Email, "deleted")
}

// levenshtein computes the edit distance between two strings (case-insensitive).
func levenshtein(a, b string) int {
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			ins := curr[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			m := ins
			if del < m {
				m = del
			}
			if sub < m {
				m = sub
			}
			curr[j] = m
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

// toTitleCase converts a string to "Title Case" (first letter of each word uppercase).
func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		runes := []rune(w)
		runes[0] = unicode.ToUpper(runes[0])
		for j := 1; j < len(runes); j++ {
			runes[j] = unicode.ToLower(runes[j])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

// fuzzyMatchJobRole finds the best matching job role using Levenshtein distance.
func fuzzyMatchJobRole(allRoles []gorm_model.JobRole, input string, maxDistance int) (*gorm_model.JobRole, bool) {
	inputLower := strings.ToLower(strings.TrimSpace(input))
	bestDist := maxDistance + 1
	var bestMatch *gorm_model.JobRole
	for i := range allRoles {
		dist := levenshtein(inputLower, strings.ToLower(allRoles[i].Name))
		if dist < bestDist {
			bestDist = dist
			bestMatch = &allRoles[i]
		}
	}
	if bestMatch != nil && bestDist <= maxDistance {
		return bestMatch, true
	}
	return nil, false
}

// getOrCreateUndefinedSector fetches or creates the "undefined" sector.
func getOrCreateUndefinedSector(db *gorm.DB) (*gorm_model.Sector, error) {
	var sector gorm_model.Sector
	if err := db.Where("LOWER(name) = ?", "undefined").First(&sector).Error; err == nil {
		return &sector, nil
	}
	sector = gorm_model.Sector{
		ID:       uuid.New().String(),
		Name:     "Undefined",
		IsActive: true,
	}
	if err := db.Create(&sector).Error; err != nil {
		return nil, err
	}
	return &sector, nil
}

// getOrCreateSectorByName fetches or creates a sector by its name (case-insensitive).
func getOrCreateSectorByName(db *gorm.DB, name string) (*gorm_model.Sector, error) {
	var sector gorm_model.Sector
	if err := db.Where("LOWER(name) = ?", strings.ToLower(name)).First(&sector).Error; err == nil {
		return &sector, nil
	}
	sector = gorm_model.Sector{
		ID:       uuid.New().String(),
		Name:     toTitleCase(name),
		IsActive: true,
	}
	if err := db.Create(&sector).Error; err != nil {
		return nil, err
	}
	return &sector, nil
}
