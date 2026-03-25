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

type SubscribeRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UpdatePreferencesRequest struct {
	Preferences string `json:"preferences"` // JSON string
}

type SubscriberResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Status      string `json:"status"`
	Preferences string `json:"preferences"`
}

type SubscriberService struct {
	repo        *postgres.SubscriberRepository
	pool        *pgxpool.Pool
	outboxStore outbox.Store
	auditLogger *audit.Logger
}

func NewSubscriberService(repo *postgres.SubscriberRepository, pool *pgxpool.Pool, outboxStore outbox.Store, auditLogger *audit.Logger) *SubscriberService {
	return &SubscriberService{repo: repo, pool: pool, outboxStore: outboxStore, auditLogger: auditLogger}
}

func (s *SubscriberService) Subscribe(ctx context.Context, req SubscribeRequest) (*SubscriberResponse, error) {
	subscriber, err := domain.NewSubscriber(req.Email, req.FirstName, req.LastName)
	if err != nil {
		return nil, err
	}

	events := subscriber.PullEvents()
	entries, err := outbox.FromDomainEvents(events, "crm.")
	if err != nil {
		return nil, fmt.Errorf("convert events to outbox: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.SaveInTx(ctx, tx, subscriber); err != nil {
		return nil, fmt.Errorf("save subscriber: %w", err)
	}

	if len(entries) > 0 {
		if err := s.outboxStore.InsertInTx(ctx, tx, entries); err != nil {
			return nil, fmt.Errorf("insert outbox: %w", err)
		}
	}

	changes, _ := json.Marshal(map[string]string{"email": subscriber.Email})
	if err := s.auditLogger.LogInTx(ctx, tx, audit.LogEntry{
		Action:     "create",
		EntityType: "subscriber",
		EntityID:   subscriber.ID,
		UserID:     "",
		Changes:    changes,
	}); err != nil {
		return nil, fmt.Errorf("audit log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return toSubscriberResponse(subscriber), nil
}

func (s *SubscriberService) GetSubscriber(ctx context.Context, id string) (*SubscriberResponse, error) {
	subscriber, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toSubscriberResponse(subscriber), nil
}

func (s *SubscriberService) Unsubscribe(ctx context.Context, id string) (*SubscriberResponse, error) {
	subscriber, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := subscriber.Unsubscribe(); err != nil {
		return nil, err
	}

	events := subscriber.PullEvents()
	entries, err := outbox.FromDomainEvents(events, "crm.")
	if err != nil {
		return nil, fmt.Errorf("convert events to outbox: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.UpdateInTx(ctx, tx, subscriber); err != nil {
		return nil, fmt.Errorf("update subscriber: %w", err)
	}

	if len(entries) > 0 {
		if err := s.outboxStore.InsertInTx(ctx, tx, entries); err != nil {
			return nil, fmt.Errorf("insert outbox: %w", err)
		}
	}

	changes, _ := json.Marshal(map[string]string{"status": "unsubscribed"})
	if err := s.auditLogger.LogInTx(ctx, tx, audit.LogEntry{
		Action:     "status_change",
		EntityType: "subscriber",
		EntityID:   subscriber.ID,
		UserID:     "",
		Changes:    changes,
	}); err != nil {
		return nil, fmt.Errorf("audit log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return toSubscriberResponse(subscriber), nil
}

func (s *SubscriberService) UpdatePreferences(ctx context.Context, id string, req UpdatePreferencesRequest) (*SubscriberResponse, error) {
	subscriber, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := subscriber.UpdatePreferences(req.Preferences); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, subscriber); err != nil {
		return nil, fmt.Errorf("update subscriber: %w", err)
	}

	return toSubscriberResponse(subscriber), nil
}

func (s *SubscriberService) ListSubscribers(ctx context.Context, filter domain.SubscriberFilter) ([]*SubscriberResponse, int, error) {
	subscribers, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*SubscriberResponse, len(subscribers))
	for i, sub := range subscribers {
		responses[i] = toSubscriberResponse(sub)
	}
	return responses, total, nil
}

func toSubscriberResponse(sub *domain.Subscriber) *SubscriberResponse {
	return &SubscriberResponse{
		ID:          sub.ID,
		Email:       sub.Email,
		FirstName:   sub.FirstName,
		LastName:    sub.LastName,
		Status:      string(sub.Status),
		Preferences: sub.Preferences,
	}
}
