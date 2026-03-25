package service

import (
	"context"
	"fmt"

	"github.com/crm-system-new/crm-marketing/internal/domain"
)

type CreateSegmentRequest struct {
	Name     string `json:"name"`
	Criteria string `json:"criteria"`
}

type UpdateSegmentRequest struct {
	Name     string `json:"name"`
	Criteria string `json:"criteria"`
}

type SegmentResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Criteria        string `json:"criteria"`
	SubscriberCount int    `json:"subscriber_count"`
}

type SegmentService struct {
	repo domain.SegmentRepository
}

func NewSegmentService(repo domain.SegmentRepository) *SegmentService {
	return &SegmentService{repo: repo}
}

func (s *SegmentService) CreateSegment(ctx context.Context, req CreateSegmentRequest) (*SegmentResponse, error) {
	segment, err := domain.NewSegment(req.Name, req.Criteria)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, segment); err != nil {
		return nil, fmt.Errorf("save segment: %w", err)
	}

	return toSegmentResponse(segment), nil
}

func (s *SegmentService) GetSegment(ctx context.Context, id string) (*SegmentResponse, error) {
	segment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toSegmentResponse(segment), nil
}

func (s *SegmentService) UpdateSegment(ctx context.Context, id string, req UpdateSegmentRequest) (*SegmentResponse, error) {
	segment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := segment.UpdateDetails(req.Name, req.Criteria); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, segment); err != nil {
		return nil, fmt.Errorf("update segment: %w", err)
	}

	return toSegmentResponse(segment), nil
}

func (s *SegmentService) ListSegments(ctx context.Context, filter domain.SegmentFilter) ([]*SegmentResponse, int, error) {
	segments, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*SegmentResponse, len(segments))
	for i, seg := range segments {
		responses[i] = toSegmentResponse(seg)
	}
	return responses, total, nil
}

func toSegmentResponse(s *domain.Segment) *SegmentResponse {
	return &SegmentResponse{
		ID:              s.ID,
		Name:            s.Name,
		Criteria:        s.Criteria,
		SubscriberCount: s.SubscriberCount,
	}
}
