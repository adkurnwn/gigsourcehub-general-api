package system_setting

import (
	"context"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
)

func (u *appUsecase) Fetch(ctx context.Context) (*gorm_model.SystemSetting, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	return u.repo.GetSystemSetting(ctx)
}

func (u *appUsecase) Update(ctx context.Context, req request_model.UpdateSystemSettingRequest) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	model, err := u.repo.GetSystemSetting(ctx)
	if err != nil {
		return err
	}

	if req.IsAIModeEnabled != nil {
		model.IsAIModeEnabled = *req.IsAIModeEnabled
	}

	return u.repo.UpdateSystemSetting(ctx, model)
}
