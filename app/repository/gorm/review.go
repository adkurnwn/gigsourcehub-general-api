package gormrepo

import (
	"context"
	"fmt"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"gorm.io/gorm"
)

// GetReviewScoresByUserIDs fetches AVG Likert scores from the reviews table
// for the given list of candidate_user_id values. Candidates with no reviews
// are simply absent from the returned map.
func (r *gormRepo) GetReviewScoresByUserIDs(ctx context.Context, userIDs []string) (map[string]domain.ReviewAggregateScore, error) {
	if len(userIDs) == 0 {
		return map[string]domain.ReviewAggregateScore{}, nil
	}

	type row struct {
		CandidateUserID             string   `gorm:"column:candidate_user_id"`
		AvgWorkQuality              *float64 `gorm:"column:avg_work_quality"`
		AvgTimeliness               *float64 `gorm:"column:avg_timeliness"`
		AvgCommunicationCollab      *float64 `gorm:"column:avg_communication_collaboration"`
		AvgProblemSolvingInitiative *float64 `gorm:"column:avg_problem_solving_initiative"`
	}

	var rows []row
	err := r.db.WithContext(ctx).
		Table("reviews r").
		Select(
			"r.candidate_user_id, " +
				"AVG(CASE WHEN q.indicator = 'WORK_QUALITY' THEN a.score END) AS avg_work_quality, " +
				"AVG(CASE WHEN q.indicator = 'TIMELINESS' THEN a.score END) AS avg_timeliness, " +
				"AVG(CASE WHEN q.indicator = 'COMMUNICATION_COLLABORATION' THEN a.score END) AS avg_communication_collaboration, " +
				"AVG(CASE WHEN q.indicator = 'PROBLEM_SOLVING_INITIATIVE' THEN a.score END) AS avg_problem_solving_initiative",
		).
		Joins("JOIN review_answers a ON a.review_id = r.id AND a.deleted_at IS NULL").
		Joins("JOIN review_questions q ON q.id = a.question_id AND q.deleted_at IS NULL").
		Where("r.candidate_user_id IN (?)", userIDs).
		Where("r.deleted_at IS NULL").
		Group("r.candidate_user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("GetReviewScoresByUserIDs query error: %w", err)
	}

	result := make(map[string]domain.ReviewAggregateScore, len(rows))
	for _, r := range rows {
		result[r.CandidateUserID] = domain.ReviewAggregateScore{
			UserID:                      r.CandidateUserID,
			AvgWorkQuality:              r.AvgWorkQuality,
			AvgTimeliness:               r.AvgTimeliness,
			AvgCommunicationCollab:      r.AvgCommunicationCollab,
			AvgProblemSolvingInitiative: r.AvgProblemSolvingInitiative,
		}
	}
	return result, nil
}

func (r *gormRepo) GetReviewByID(ctx context.Context, id string) (*gorm_model.Review, error) {
	var review gorm_model.Review
	err := r.db.WithContext(ctx).
		Preload("Answers").
		Preload("Answers.Question").
		Where("id = ?", id).
		First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *gormRepo) GetReviewByOnboardEmployee(ctx context.Context, onboardHistoryID, employeeUserID string) (*gorm_model.Review, error) {
	var review gorm_model.Review
	err := r.db.WithContext(ctx).
		Where("onboard_history_id = ? AND employee_user_id = ?", onboardHistoryID, employeeUserID).
		First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *gormRepo) CreateReview(ctx context.Context, model *gorm_model.Review) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) UpdateReview(ctx context.Context, model *gorm_model.Review) error {
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *gormRepo) FetchReviewQuestions(ctx context.Context) ([]gorm_model.ReviewQuestion, error) {
	var rows []gorm_model.ReviewQuestion
	err := r.db.WithContext(ctx).
		Order("indicator ASC, question_order ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *gormRepo) ReplaceReviewAnswers(ctx context.Context, reviewID string, answers []gorm_model.ReviewAnswer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("review_id = ?", reviewID).Delete(&gorm_model.ReviewAnswer{}).Error; err != nil {
			return err
		}
		if len(answers) == 0 {
			return nil
		}
		return tx.Create(&answers).Error
	})
}
