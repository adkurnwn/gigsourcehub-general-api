package usecase_member

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func (u *appUsecase) FetchUsers(ctx context.Context, page, limit int64, cursor string, roleName *string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit

	// Calculate totals based on a raw manual query joined explicitly against role system to filter the counter safely before fetching items
	db := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.User{})

	if roleName != nil {
		db = db.Joins("JOIN role_systems rs ON users.role_system_id = rs.id").Where("rs.name = ?", *roleName)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count users")
	}

	// Execute actual limited fetch
	var users []gorm_model.User
	if err := db.Preload("RoleSystem").Limit(int(limit)).Offset(int(offset)).Order("created_at DESC").Find(&users).Error; err != nil {
		logrus.Error("FetchUsers error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch users")
	}

	var results []interface{}
	for _, user := range users {
		results = append(results, user.ToUserResp())
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
		user.RoleSystem = &gorm_model.RoleSystem{
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
		user.RoleSystem = &gorm_model.RoleSystem{
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

	user := gorm_model.User{
		ID:             uuid.New().String(),
		Email:          req.Email,
		Name:           req.Name,
		Password:       string(hashedPassword),
		AssignedRoleId: req.AssignedRoleId,
		RoleSystemId:   req.RoleSystemId,
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

	if user.RoleSystem.Name == "Admin" || user.RoleSystem.Name == "Employee" || user.RoleSystem.Name == "Superadmin" {
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

	if user.RoleSystem.Name == "Candidate" {
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
