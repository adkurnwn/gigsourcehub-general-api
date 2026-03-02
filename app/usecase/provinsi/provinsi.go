package usecase_provinsi

import (
	"context"
	"net/http"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/sirupsen/logrus"
)

func (u *appUsecase) FetchAll(ctx context.Context, filter gorm_model.ProvinsiFilter) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if len(filter.Sorts) == 0 {
		filter.Sorts = []map[string]string{
			{"created_at": "DESC"},
		}
	}

	// Query Database for Paginated Rows
	rows, err := u.gormDbRepo.FetchProvinsi(ctx, filter)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch Provinsi data")
	}
	defer rows.Close()

	// Parse database response cursors
	var results []interface{}
	for rows.Next() {
		var provinsi gorm_model.Provinsi
		if err := u.gormDbRepo.StructScan(rows, &provinsi); err != nil {
			logrus.Error("Provinsi map error:", err)
			continue
		}
		results = append(results, provinsi.ToProvinsiResp())
	}

	if results == nil {
		results = []interface{}{}
	}

	return response.Success(map[string]interface{}{
		"list": results,
	})
}

func (u *appUsecase) FetchData(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	data, err := u.gormDbRepo.FetchProvinsi(ctx, gorm_model.ProvinsiFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch Provinsi")
	}
	defer data.Close()

	if !data.Next() {
		return response.Error(http.StatusNotFound, "Provinsi not found")
	}

	var provinsi gorm_model.Provinsi
	if err := u.gormDbRepo.StructScan(data, &provinsi); err != nil {
		logrus.Error("Provinsi struct map error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to serialize Provinsi data")
	}

	return response.Success(provinsi.ToProvinsiResp())
}
