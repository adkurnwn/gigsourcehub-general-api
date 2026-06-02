package usecase_bookmark

import (
	"context"
	"net/http"
	"strconv"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/sirupsen/logrus"
)

func (u *appUsecase) Create(ctx context.Context, adminID string, req request_model.CreateBookmarkRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Verify the candidate exists
	candidateUser, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: req.CandidateID},
	})
	if err != nil || candidateUser == nil {
		return response.Error(http.StatusNotFound, "Candidate user not found")
	}

	// Check if bookmark already exists
	existing, _ := u.gormDbRepo.GetBookmark(ctx, adminID, req.CandidateID)
	if existing != nil {
		return response.Error(http.StatusConflict, "Bookmark already created!")
	}

	bookmark := gorm_model.Bookmark{
		AdminID:     adminID,
		CandidateID: req.CandidateID,
	}

	if err := u.gormDbRepo.CreateBookmark(ctx, &bookmark); err != nil {
		logrus.Error("Bookmark Create error: ", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Bookmark", "Candidate", req.CandidateID, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to create bookmark")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Bookmark", "Candidate", req.CandidateID, req, true)

	return response.Success(bookmark)
}

func (u *appUsecase) Delete(ctx context.Context, adminID string, candidateID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	rowsAffected, err := u.gormDbRepo.DeleteBookmark(ctx, adminID, candidateID)
	if err != nil {
		logrus.Error("Bookmark Delete error: ", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Unbookmark", "Candidate", candidateID, nil, false)
		return response.Error(http.StatusInternalServerError, "Failed to delete bookmark")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Unbookmark", "Candidate", candidateID, nil, true)

	if rowsAffected == 0 {
		return response.Error(http.StatusNotFound, "Bookmark already deleted!")
	}

	return response.Success(nil)
}

func (u *appUsecase) FetchByAdmin(ctx context.Context, adminID string, page, limit int64) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit

	// Count total
	total, err := u.gormDbRepo.CountBookmarksByAdmin(ctx, adminID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count bookmarks")
	}

	// Fetch paginated rows
	rows, err := u.gormDbRepo.FetchBookmarksByAdmin(ctx, adminID, limit, offset)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch bookmarks")
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		var bookmark gorm_model.Bookmark
		if err := u.gormDbRepo.StructScan(rows, &bookmark); err != nil {
			logrus.Error("Bookmark scan error: ", err)
			continue
		}
		results = append(results, bookmark)
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
