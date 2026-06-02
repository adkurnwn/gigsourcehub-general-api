package usecase_kabupaten_kota

import (
	"context"
	"net/http"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/sirupsen/logrus"
)

func (u *appUsecase) FetchAll(ctx context.Context, filter gorm_model.KabupatenKotaFilter) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if len(filter.Sorts) == 0 {
		filter.Sorts = []map[string]string{
			{"created_at": "DESC"},
		}
	}

	// Query Database for Paginated Rows
	rows, err := u.gormDbRepo.FetchKabupatenKota(ctx, filter)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch KabupatenKota data")
	}
	defer rows.Close()

	// Parse database response cursors
	var results []interface{}
	for rows.Next() {
		var kabupatenKota gorm_model.KabupatenKota
		if err := u.gormDbRepo.StructScan(rows, &kabupatenKota); err != nil {
			logrus.Error("KabupatenKota map error:", err)
			continue
		}
		results = append(results, kabupatenKota.ToKabupatenKotaResp())
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

	data, err := u.gormDbRepo.FetchKabupatenKota(ctx, gorm_model.KabupatenKotaFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch KabupatenKota")
	}
	defer data.Close()

	if !data.Next() {
		return response.Error(http.StatusNotFound, "KabupatenKota not found")
	}

	var kabupatenKota gorm_model.KabupatenKota
	if err := u.gormDbRepo.StructScan(data, &kabupatenKota); err != nil {
		logrus.Error("KabupatenKota struct map error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to serialize KabupatenKota data")
	}

	return response.Success(kabupatenKota.ToKabupatenKotaResp())
}
