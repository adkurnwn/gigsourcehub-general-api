package grpc

import (
	"context"

	usecase_cv "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/cv"
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

func (h *CandidateHandler) GetSectors(ctx context.Context, req *pb.GetSectorsRequest) (*pb.GetSectorsResponse, error) {
	sectors, err := h.gormRepo.GetActiveSectors(ctx)
	if err != nil {
		return nil, err
	}

	var pbSectors []*pb.SectorInfo
	for _, s := range sectors {
		pbSectors = append(pbSectors, &pb.SectorInfo{
			Id:   s.ID,
			Name: s.Name,
		})
	}

	return &pb.GetSectorsResponse{
		Sectors: pbSectors,
	}, nil
}

func (h *CandidateHandler) GetJobRoles(ctx context.Context, req *pb.GetJobRolesRequest) (*pb.GetJobRolesResponse, error) {
	roles, err := h.gormRepo.GetActiveJobRoles(ctx)
	if err != nil {
		return nil, err
	}

	var pbRoles []*pb.JobRoleInfo
	for _, r := range roles {
		pbRoles = append(pbRoles, &pb.JobRoleInfo{
			Id:   r.ID,
			Name: r.Name,
		})
	}

	return &pb.GetJobRolesResponse{
		JobRoles: pbRoles,
	}, nil
}

func (h *CandidateHandler) UpdateCVProgress(ctx context.Context, req *pb.UpdateCVProgressRequest) (*pb.UpdateCVProgressResponse, error) {
	cvID := req.GetCvId()
	progress := req.GetProgress()
	status := req.GetStatus()

	cv, err := h.gormRepo.GetCVByID(ctx, cvID)
	if err != nil {
		return &pb.UpdateCVProgressResponse{
			Success: false,
			Message: "cv not found: " + err.Error(),
		}, nil
	}

	if status != "" {
		cv.Status = status
		if err := h.gormRepo.UpdateCV(ctx, cv); err != nil {
			return &pb.UpdateCVProgressResponse{
				Success: false,
				Message: "failed to update cv status: " + err.Error(),
			}, nil
		}
	}

	usecase_cv.SetCVProgress(cvID, int(progress))

	return &pb.UpdateCVProgressResponse{
		Success: true,
		Message: "success",
	}, nil
}
