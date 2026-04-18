package usecase_member

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	pb "github.com/adkurnwn/gigsourcehub-general-api/proto"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func (u *appUsecase) FetchUsers(ctx context.Context, page, limit int64, cursor string, roleName *string, adminID *string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit

	// Calculate totals based on a raw manual query joined explicitly against role system to filter the counter safely before fetching items
	db := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.User{})

	if roleName != nil {
		db = db.Joins("JOIN system_roles rs ON users.system_role_id = rs.id").Where("rs.name = ?", *roleName)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count users")
	}

	// Execute actual limited fetch
	var users []gorm_model.User
	if err := db.Preload("SystemRole").
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
			ID:   cv.ID,
			Name: cv.Filename,
			URL:  presignedLink,
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

	user := gorm_model.User{
		ID:             uuid.New().String(),
		Email:          req.Email,
		Name:           req.Name,
		Password:       string(hashedPassword),
		AssignedRoleId: req.AssignedRoleId,
		SystemRoleId:   req.SystemRoleId,
		JobTitleId:     req.JobTitleId,
	}

	if req.AccountStatus != nil {
		user.AccountStatus = req.AccountStatus
	}

	if err := u.gormDbRepo.CreateUser(ctx, &user); err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}

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

	if req.Name != nil && *req.Name != "" {
		user.Name = *req.Name
	}

	if req.AssignedRoleId != nil {
		user.AssignedRoleId = req.AssignedRoleId
	}

	if req.AccountStatus != nil {
		user.AccountStatus = req.AccountStatus
	}

	if err := u.gormDbRepo.UpdateUser(ctx, user); err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}

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
		return response.Error(http.StatusInternalServerError, err.Error())
	}

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
		return response.Error(http.StatusInternalServerError, err.Error())
	}

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
		return response.Error(http.StatusInternalServerError, err.Error())
	}

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

	if req.JobRoleIds != nil {
		var jobRoles []gorm_model.JobRole
		if len(req.JobRoleIds) > 0 {
			if err := u.gormDbRepo.GetDB().WithContext(ctx).Where("id IN ?", req.JobRoleIds).Find(&jobRoles).Error; err != nil {
				logrus.Errorf("failed to fetch job roles: %v", err)
			}
		}

		if err := u.gormDbRepo.GetDB().WithContext(ctx).Model(user).Association("JobRoles").Replace(jobRoles); err != nil {
			logrus.Errorf("failed to update job roles association: %v", err)
		}

		user.JobRoles = jobRoles
	}

	if err := u.gormDbRepo.UpdateUser(ctx, user); err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}

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
