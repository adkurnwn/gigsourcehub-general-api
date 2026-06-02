package usecase_review

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
)

type reviewSnapshot struct {
	SubrequestID   string `json:"subrequest_id"`
	EmployeeUserID string `json:"employee_user_id"`
}

func (u *appUsecase) CreateOrUpdate(ctx context.Context, employeeID string, req request_model.CreateReviewRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if employeeID == "" {
		return response.Error(http.StatusBadRequest, "employee_id is required")
	}
	if req.OnboardHistoryID == "" {
		return response.Error(http.StatusBadRequest, "onboard_history_id is required")
	}

	onboard, err := u.gormDbRepo.GetOnboardHistoryByID(ctx, req.OnboardHistoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.Error(http.StatusNotFound, "Onboarding history not found")
		}
		return response.Error(http.StatusInternalServerError, "Failed to fetch onboarding history")
	}

	now := time.Now().UTC()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if onboard.EndDate == nil || onboard.EndDate.After(currentDate) {
		return response.Error(http.StatusBadRequest, "Onboarding is not finished")
	}

	if onboard.Snapshot == nil || *onboard.Snapshot == "" {
		return response.Error(http.StatusBadRequest, "Onboarding snapshot is missing")
	}

	var snapshot reviewSnapshot
	if err := json.Unmarshal([]byte(*onboard.Snapshot), &snapshot); err != nil {
		return response.Error(http.StatusBadRequest, "Invalid onboarding snapshot")
	}
	if snapshot.SubrequestID == "" || snapshot.EmployeeUserID == "" {
		return response.Error(http.StatusBadRequest, "Onboarding snapshot is incomplete")
	}
	if snapshot.EmployeeUserID != employeeID {
		return response.Error(http.StatusForbidden, "You are not assigned to this onboarding")
	}

	answersByIndicator := map[string][]int{
		"WORK_QUALITY":                req.WorkQuality,
		"TIMELINESS":                  req.Timeliness,
		"COMMUNICATION_COLLABORATION": req.CommunicationCollaboration,
		"PROBLEM_SOLVING_INITIATIVE":  req.ProblemSolvingInitiative,
	}

	for indicator, scores := range answersByIndicator {
		if len(scores) != 4 {
			return response.Error(http.StatusBadRequest, "Each indicator must have exactly 4 answers")
		}
		for _, score := range scores {
			if score < 1 || score > 5 {
				return response.Error(http.StatusBadRequest, "Scores must be between 1 and 5")
			}
		}
		if indicator == "" {
			return response.Error(http.StatusBadRequest, "Indicator is required")
		}
	}

	questions, err := u.gormDbRepo.FetchReviewQuestions(ctx)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to load review questions")
	}

	questionsByIndicator := map[string][]gorm_model.ReviewQuestion{}
	for _, q := range questions {
		questionsByIndicator[q.Indicator] = append(questionsByIndicator[q.Indicator], q)
	}

	for indicator := range answersByIndicator {
		if len(questionsByIndicator[indicator]) != 4 {
			return response.Error(http.StatusInternalServerError, "Review questions are not configured correctly")
		}
	}

	var answers []gorm_model.ReviewAnswer
	for indicator, scores := range answersByIndicator {
		indicatorQuestions := questionsByIndicator[indicator]
		for i, score := range scores {
			answers = append(answers, gorm_model.ReviewAnswer{
				QuestionID: indicatorQuestions[i].ID,
				Score:      score,
			})
		}
	}

	action := "Create"
	reviewID := ""

	existing, err := u.gormDbRepo.GetReviewByOnboardEmployee(ctx, req.OnboardHistoryID, employeeID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return response.Error(http.StatusInternalServerError, "Failed to check review")
	}

	if existing == nil {
		review := gorm_model.Review{
			SubrequestID:     snapshot.SubrequestID,
			CandidateUserID:  onboard.CandidateUserID,
			EmployeeUserID:   employeeID,
			OnboardHistoryID: req.OnboardHistoryID,
			FinalRecommendation: func() string {
				if req.FinalRecommendation != nil {
					return *req.FinalRecommendation
				}
				return ""
			}(),
			Notes: req.Notes,
		}

		if err := u.gormDbRepo.CreateReview(ctx, &review); err != nil {
			helpers.LogActivity(ctx, u.gormDbRepo, "Create", "Review", req.OnboardHistoryID, req, false)
			return response.Error(http.StatusInternalServerError, "Failed to create review")
		}
		reviewID = review.ID
	} else {
		action = "Update"
		existing.FinalRecommendation = ""
		if req.FinalRecommendation != nil {
			existing.FinalRecommendation = *req.FinalRecommendation
		}
		existing.Notes = req.Notes
		if err := u.gormDbRepo.UpdateReview(ctx, existing); err != nil {
			helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Review", existing.ID, req, false)
			return response.Error(http.StatusInternalServerError, "Failed to update review")
		}
		reviewID = existing.ID
	}

	for i := range answers {
		answers[i].ReviewID = reviewID
	}

	if err := u.gormDbRepo.ReplaceReviewAnswers(ctx, reviewID, answers); err != nil {
		helpers.LogActivity(ctx, u.gormDbRepo, action, "Review", reviewID, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to save review answers")
	}

	review, err := u.gormDbRepo.GetReviewByID(ctx, reviewID)
	if err != nil {
		helpers.LogActivity(ctx, u.gormDbRepo, action, "Review", reviewID, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to load review")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, action, "Review", reviewID, req, true)
	return response.Success(review.ToReviewResp())
}

func (u *appUsecase) FetchByID(ctx context.Context, reviewID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if reviewID == "" {
		return response.Error(http.StatusBadRequest, "review_id is required")
	}

	review, err := u.gormDbRepo.GetReviewByID(ctx, reviewID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.Error(http.StatusNotFound, "Review not found")
		}
		logrus.Errorf("Review FetchByID error: %v", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch review")
	}

	return response.Success(review.ToReviewResp())
}

func (u *appUsecase) FetchQuestions(ctx context.Context) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	questions, err := u.gormDbRepo.FetchReviewQuestions(ctx)
	if err != nil {
		logrus.Errorf("Review FetchQuestions error: %v", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch review questions")
	}

	results := make([]gorm_model.ReviewQuestionResp, 0, len(questions))
	for i := range questions {
		results = append(results, questions[i].ToReviewQuestionResp())
	}

	return response.Success(results)
}
