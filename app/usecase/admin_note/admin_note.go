package usecase_admin_note

import (
	"context"
	"net/http"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/sirupsen/logrus"
)

func (u *appUsecase) Create(ctx context.Context, actorID string, candidateID string, req request_model.CreateAdminNoteRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	candidateUser, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: candidateID},
	})
	if err != nil || candidateUser == nil {
		return response.Error(http.StatusNotFound, "Candidate user not found")
	}

	note := gorm_model.AdminNote{
		AdminUserID:     actorID,
		CandidateUserID: candidateID,
		Content:         req.Content,
	}

	if err := u.gormDbRepo.CreateAdminNote(ctx, &note); err != nil {
		logrus.Error("AdminNote Create error: ", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Add Note", "Candidate", candidateID, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to create note")
	}

	adminUser, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: actorID},
	})
	if err != nil {
		logrus.Error("AdminNote Fetch admin error: ", err)
	} else {
		note.AdminUser = adminUser
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Add Note", "Candidate", candidateID, req, true)
	return response.Success(note.ToAdminNoteResp())
}

func (u *appUsecase) FetchByCandidate(ctx context.Context, candidateID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	candidateUser, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: candidateID},
	})
	if err != nil || candidateUser == nil {
		return response.Error(http.StatusNotFound, "Candidate user not found")
	}

	notes, err := u.gormDbRepo.FetchAdminNotesByCandidate(ctx, candidateID)
	if err != nil {
		logrus.Error("AdminNote Fetch error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch notes")
	}

	results := make([]gorm_model.AdminNoteResp, 0, len(notes))
	for i := range notes {
		results = append(results, notes[i].ToAdminNoteResp())
	}

	return response.Success(results)
}
