package grpc

import (
	"context"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	pb "github.com/adkurnwn/gigsourcehub-general-api/proto"
)

type CandidateHandler struct {
	pb.UnimplementedCandidateServiceServer
	gormRepo domain.GormRepo
}

func NewCandidateHandler(repo domain.GormRepo) *CandidateHandler {
	return &CandidateHandler{
		gormRepo: repo,
	}
}

func (h *CandidateHandler) GetEnrichmentData(ctx context.Context, req *pb.EnrichmentRequest) (*pb.EnrichmentResponse, error) {
	userIDs := req.GetUserIds()
	if len(userIDs) == 0 {
		return &pb.EnrichmentResponse{}, nil
	}

	scores, err := h.gormRepo.GetReviewScoresByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	jobRoles, err := h.gormRepo.GetJobRolesByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	candidateLevels, err := h.gormRepo.GetCandidateLevelsByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	var data []*pb.EnrichmentData
	for _, id := range userIDs {
		s, hasScore := scores[id]
		roles, _ := jobRoles[id]
		level := candidateLevels[id]

		item := &pb.EnrichmentData{
			UserId:         id,
			JobRoles:       roles,
			CandidateLevel: level,
			HasScore:       hasScore,
		}

		if hasScore {
			if s.AvgWorkQuality != nil {
				item.AvgWorkQuality = *s.AvgWorkQuality
			}
			if s.AvgTimeliness != nil {
				item.AvgTimeliness = *s.AvgTimeliness
			}
			if s.AvgCommunicationCollab != nil {
				item.AvgCommunicationCollaboration = *s.AvgCommunicationCollab
			}
			if s.AvgProblemSolvingInitiative != nil {
				item.AvgProblemSolvingInitiative = *s.AvgProblemSolvingInitiative
			}
		}

		data = append(data, item)
	}

	return &pb.EnrichmentResponse{
		Data: data,
	}, nil
}
