package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/crm-system-new/crm-marketing/internal/domain"
	"github.com/crm-system-new/crm-marketing/internal/infrastructure/postgres"
	"github.com/crm-system-new/crm-shared/pkg/audit"
	"github.com/crm-system-new/crm-shared/pkg/outbox"
)

type CreateCampaignRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Channel     string  `json:"channel"`
	SegmentID   string  `json:"segment_id"`
	ScheduledAt *string `json:"scheduled_at,omitempty"` // RFC3339
}

type CampaignResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Channel     string  `json:"channel"`
	SegmentID   string  `json:"segment_id"`
	ScheduledAt *string `json:"scheduled_at,omitempty"`
	SentCount   int     `json:"sent_count"`
	OpenRate    float64 `json:"open_rate"`
	ClickRate   float64 `json:"click_rate"`
}

type CampaignService struct {
	repo        *postgres.CampaignRepository
	pool        *pgxpool.Pool
	outboxStore outbox.Store
	auditLogger *audit.Logger
}

func NewCampaignService(repo *postgres.CampaignRepository, pool *pgxpool.Pool, outboxStore outbox.Store, auditLogger *audit.Logger) *CampaignService {
	return &CampaignService{repo: repo, pool: pool, outboxStore: outboxStore, auditLogger: auditLogger}
}

func (s *CampaignService) CreateCampaign(ctx context.Context, req CreateCampaignRequest) (*CampaignResponse, error) {
	var scheduledAt *time.Time
	if req.ScheduledAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err == nil {
			scheduledAt = &t
		}
	}

	campaign, err := domain.NewCampaign(req.Name, req.Description, req.Channel, req.SegmentID, scheduledAt)
	if err != nil {
		return nil, err
	}

	events := campaign.PullEvents()
	entries, err := outbox.FromDomainEvents(events, "crm.")
	if err != nil {
		return nil, fmt.Errorf("convert events to outbox: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.SaveInTx(ctx, tx, campaign); err != nil {
		return nil, fmt.Errorf("save campaign: %w", err)
	}

	if len(entries) > 0 {
		if err := s.outboxStore.InsertInTx(ctx, tx, entries); err != nil {
			return nil, fmt.Errorf("insert outbox: %w", err)
		}
	}

	changes, _ := json.Marshal(map[string]string{"name": campaign.Name, "channel": campaign.Channel})
	if err := s.auditLogger.LogInTx(ctx, tx, audit.LogEntry{
		Action:     "create",
		EntityType: "campaign",
		EntityID:   campaign.ID,
		UserID:     "",
		Changes:    changes,
	}); err != nil {
		return nil, fmt.Errorf("audit log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return toCampaignResponse(campaign), nil
}

func (s *CampaignService) GetCampaign(ctx context.Context, id string) (*CampaignResponse, error) {
	campaign, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toCampaignResponse(campaign), nil
}

func (s *CampaignService) LaunchCampaign(ctx context.Context, id string) (*CampaignResponse, error) {
	campaign, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := campaign.Launch(); err != nil {
		return nil, err
	}

	events := campaign.PullEvents()
	entries, err := outbox.FromDomainEvents(events, "crm.")
	if err != nil {
		return nil, fmt.Errorf("convert events to outbox: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.UpdateInTx(ctx, tx, campaign); err != nil {
		return nil, fmt.Errorf("update campaign: %w", err)
	}

	if len(entries) > 0 {
		if err := s.outboxStore.InsertInTx(ctx, tx, entries); err != nil {
			return nil, fmt.Errorf("insert outbox: %w", err)
		}
	}

	changes, _ := json.Marshal(map[string]string{"action": "launch"})
	if err := s.auditLogger.LogInTx(ctx, tx, audit.LogEntry{
		Action:     "status_change",
		EntityType: "campaign",
		EntityID:   campaign.ID,
		UserID:     "",
		Changes:    changes,
	}); err != nil {
		return nil, fmt.Errorf("audit log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return toCampaignResponse(campaign), nil
}

func (s *CampaignService) PauseCampaign(ctx context.Context, id string) (*CampaignResponse, error) {
	campaign, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := campaign.Pause(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, campaign); err != nil {
		return nil, fmt.Errorf("update campaign: %w", err)
	}

	return toCampaignResponse(campaign), nil
}

func (s *CampaignService) ListCampaigns(ctx context.Context, filter domain.CampaignFilter) ([]*CampaignResponse, int, error) {
	campaigns, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*CampaignResponse, len(campaigns))
	for i, c := range campaigns {
		responses[i] = toCampaignResponse(c)
	}
	return responses, total, nil
}

func toCampaignResponse(c *domain.Campaign) *CampaignResponse {
	resp := &CampaignResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Status:      string(c.Status),
		Channel:     c.Channel,
		SegmentID:   c.SegmentID,
		SentCount:   c.SentCount,
		OpenRate:    c.OpenRate,
		ClickRate:   c.ClickRate,
	}
	if c.ScheduledAt != nil {
		s := c.ScheduledAt.Format(time.RFC3339)
		resp.ScheduledAt = &s
	}
	return resp
}
