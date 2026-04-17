package http_internal

import (
	"net/http"
	"strings"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/gin-gonic/gin"
)

// ReviewScoreResponse is the JSON shape returned to the AI API per candidate.
type ReviewScoreResponse struct {
	UserID                      string   `json:"user_id"`
	AvgWorkQuality              *float64 `json:"avg_work_quality"`
	AvgTimeliness               *float64 `json:"avg_timeliness"`
	AvgCommunicationCollab      *float64 `json:"avg_communication_collaboration"`
	AvgProblemSolvingInitiative *float64 `json:"avg_problem_solving_initiative"`
	JobRoles                    []string `json:"job_roles"`
	CandidateLevel              string   `json:"candidate_level"`
}

type InternalHandler struct {
	gormRepo domain.GormRepo
}

// NewInternalHandler registers internal routes used by other backend services (e.g. ai-api).
func NewInternalHandler(r *gin.RouterGroup, mdl middleware.Middleware, repo domain.GormRepo) {
	h := &InternalHandler{gormRepo: repo}

	internal := r.Group("/internal")
	internal.Use(mdl.AuthInternal())
	internal.GET("/reviews", h.GetReviewScores)
}

// GetReviewScores returns aggregated review scores and job roles for a comma-separated list of candidate user IDs.
// Query param: user_ids=uuid1,uuid2,...
func (h *InternalHandler) GetReviewScores(c *gin.Context) {
	rawIDs := c.Query("user_ids")
	if rawIDs == "" {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "user_ids query param is required"))
		return
	}

	userIDs := strings.Split(rawIDs, ",")
	for i, id := range userIDs {
		userIDs[i] = strings.TrimSpace(id)
	}

	scores, err := h.gormRepo.GetReviewScoresByUserIDs(c.Request.Context(), userIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(http.StatusInternalServerError, "failed to fetch review scores: "+err.Error()))
		return
	}

	jobRoles, err := h.gormRepo.GetJobRolesByUserIDs(c.Request.Context(), userIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(http.StatusInternalServerError, "failed to fetch job roles: "+err.Error()))
		return
	}

	candidateLevels, err := h.gormRepo.GetCandidateLevelsByUserIDs(c.Request.Context(), userIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(http.StatusInternalServerError, "failed to fetch candidate levels: "+err.Error()))
		return
	}

	var result []ReviewScoreResponse
	for _, id := range userIDs {
		s, hasScore := scores[id]
		roles, hasRoles := jobRoles[id]
		if !hasRoles {
			roles = []string{}
		}
		level := candidateLevels[id]

		item := ReviewScoreResponse{
			UserID:         id,
			JobRoles:       roles,
			CandidateLevel: level,
		}

		if hasScore {
			item.AvgWorkQuality = s.AvgWorkQuality
			item.AvgTimeliness = s.AvgTimeliness
			item.AvgCommunicationCollab = s.AvgCommunicationCollab
			item.AvgProblemSolvingInitiative = s.AvgProblemSolvingInitiative
		}

		result = append(result, item)
	}

	c.JSON(http.StatusOK, response.Success(result))
}
