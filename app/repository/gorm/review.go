package gormrepo

import (
	"context"
	"fmt"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
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
		Table("reviews").
		Select(
			"candidate_user_id, " +
				"AVG(work_quality) AS avg_work_quality, " +
				"AVG(timeliness) AS avg_timeliness, " +
				"AVG(communication_collaboration) AS avg_communication_collaboration, " +
				"AVG(problem_solving_initiative) AS avg_problem_solving_initiative",
		).
		Where("candidate_user_id IN (?)", userIDs).
		Where("deleted_at IS NULL").
		Group("candidate_user_id").
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
