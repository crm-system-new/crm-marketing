package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/crm-system-new/crm-marketing/internal/domain"
	"github.com/crm-system-new/crm-marketing/internal/infrastructure/postgres"
	"github.com/crm-system-new/crm-shared/pkg/audit"
	"github.com/crm-system-new/crm-shared/pkg/outbox"
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
	repo        *postgres.SegmentRepository
	pool        *pgxpool.Pool
	outboxStore outbox.Store
	auditLogger *audit.Logger
}

func NewSegmentService(repo *postgres.SegmentRepository, pool *pgxpool.Pool, outboxStore outbox.Store, auditLogger *audit.Logger) *SegmentService {
	return &SegmentService{repo: repo, pool: pool, outboxStore: outboxStore, auditLogger: auditLogger}
}

func (s *SegmentService) CreateSegment(ctx context.Context, req CreateSegmentRequest) (*SegmentResponse, error) {
	segment, err := domain.NewSegment(req.Name, req.Criteria)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.SaveInTx(ctx, tx, segment); err != nil {
		return nil, fmt.Errorf("save segment: %w", err)
	}

	changes, _ := json.Marshal(map[string]string{"name": segment.Name})
	if err := s.auditLogger.LogInTx(ctx, tx, audit.LogEntry{
		Action:     "create",
		EntityType: "segment",
		EntityID:   segment.ID,
		UserID:     "",
		Changes:    changes,
	}); err != nil {
		return nil, fmt.Errorf("audit log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
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

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.UpdateInTx(ctx, tx, segment); err != nil {
		return nil, fmt.Errorf("update segment: %w", err)
	}

	changes, _ := json.Marshal(map[string]string{"name": segment.Name, "criteria": segment.Criteria})
	if err := s.auditLogger.LogInTx(ctx, tx, audit.LogEntry{
		Action:     "update",
		EntityType: "segment",
		EntityID:   segment.ID,
		UserID:     "",
		Changes:    changes,
	}); err != nil {
		return nil, fmt.Errorf("audit log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
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
