package ai_search

import (
	"context"
	"fmt"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model"
	pb "github.com/adkurnwn/gigsourcehub-general-api/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type aiSearchRepository struct {
	client pb.SearchServiceClient
	conn   *grpc.ClientConn
}

func NewAISearchRepository(target string) (domain.AISearchRepository, error) {
	// Using grpc.WithInsecure() for compatibility with older grpc versions (v1.36.x)
	// If newer version is used, replace with credentials/insecure
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ai search service: %w", err)
	}

	client := pb.NewSearchServiceClient(conn)

	return &aiSearchRepository{
		client: client,
		conn:   conn,
	}, nil
}

func (r *aiSearchRepository) Close() error {
	return r.conn.Close()
}

func (r *aiSearchRepository) Search(ctx context.Context, query string) ([]model.SearchResult, error) {
	req := &pb.SearchRequest{
		Query: query,
	}

	logrus.Infof("[gRPC] Sending Search Request: query=%s", query)

	resp, err := r.client.Search(ctx, req)
	if err != nil {
		logrus.Errorf("[gRPC] Search Request Failed: %v", err)
		return nil, err
	}

	logrus.Infof("[gRPC] Search Request Success: found %d results", len(resp.Results))

	var results []model.SearchResult
	for _, item := range resp.Results {
		results = append(results, model.SearchResult{
			ID:      item.Id,
			Content: item.Content,
			Score:   item.Score,
		})
	}

	return results, nil
}
