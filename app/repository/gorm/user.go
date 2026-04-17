package gormrepo

import (
	"context"
	"database/sql"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (r *gormRepo) FetchUser(ctx context.Context, options gorm_model.UserFilter) (cur *sql.Rows, err error) {
	// generate query
	q := r.db.Model(&gorm_model.User{})
	options.Query(q)

	q = q.Preload("SystemRole")

	cur, err = q.WithContext(ctx).Rows()
	if err != nil {
		logrus.Error("FetchUser Find:", err)
		return
	}

	return
}

func (r *gormRepo) FetchOneUser(ctx context.Context, options gorm_model.UserFilter) (row *gorm_model.User, err error) {
	// generate query
	q := r.db.Model(&gorm_model.User{})
	options.Query(q)

	q = q.Preload("SystemRole")

	// set row
	row = new(gorm_model.User)

	err = q.WithContext(ctx).First(row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logrus.Error("FetchOneUser Get:", err)
		return
	}

	return
}

func (r *gormRepo) CountUser(ctx context.Context, options gorm_model.UserFilter) (total int64) {
	// generate query
	q := r.db.Model(&gorm_model.User{})
	options.Query(q)

	err := q.WithContext(ctx).Count(&total).Error
	if err != nil {
		logrus.Error("CountUser", err)
		return 0
	}

	return
}

func (r *gormRepo) CreateUser(ctx context.Context, row *gorm_model.User) (err error) {
	err = r.db.WithContext(ctx).Create(row).Error
	if err != nil {
		logrus.Error("CreateUser Exec:", err)
		return
	}

	return
}

func (r *gormRepo) UpdateUser(ctx context.Context, row *gorm_model.User) (err error) {
	err = r.db.WithContext(ctx).Save(row).Error
	if err != nil {
		logrus.Error("UpdateUser Exec:", err)
		return
	}

	return
}

func (r *gormRepo) GetProvinsiName(ctx context.Context, id string) (name string, err error) {
	err = r.db.WithContext(ctx).Table("provinsi").Where("id = ?", id).Select("name").Row().Scan(&name)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return
}

func (r *gormRepo) GetKabupatenName(ctx context.Context, id string) (name string, err error) {
	err = r.db.WithContext(ctx).Table("kabupaten_kota").Where("id = ?", id).Select("name").Row().Scan(&name)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return
}

func (r *gormRepo) GetRoleNameByUserID(ctx context.Context, userID string) (roleName string, err error) {
	err = r.db.WithContext(ctx).
		Table("users u").
		Joins("JOIN system_roles rs ON u.system_role_id = rs.id").
		Where("u.id = ?", userID).
		Select("rs.name").
		Row().
		Scan(&roleName)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return roleName, err
}

func (r *gormRepo) CreateUserBySuperadmin(ctx context.Context, row *gorm_model.User) (err error) {
	err = r.db.WithContext(ctx).Create(row).Error
	if err != nil {
		logrus.Error("CreateUserBySuperadmin Exec:", err)
		return
	}

	return
}

func (r *gormRepo) GetCandidateLevelsByUserIDs(ctx context.Context, userIDs []string) (map[string]string, error) {
	if len(userIDs) == 0 {
		return map[string]string{}, nil
	}

	type row struct {
		ID             string  `gorm:"column:id"`
		CandidateLevel *string `gorm:"column:candidate_level"`
	}

	var rows []row
	err := r.db.WithContext(ctx).
		Table("users").
		Select("id, candidate_level").
		Where("id IN (?)", userIDs).
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, r := range rows {
		level := ""
		if r.CandidateLevel != nil {
			level = *r.CandidateLevel
		}
		result[r.ID] = level
	}

	return result, nil
}

