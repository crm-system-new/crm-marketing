package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/crm-system-new/crm-marketing/internal/domain"
	"github.com/crm-system-new/crm-shared/pkg/ddd"
	"github.com/crm-system-new/crm-shared/pkg/messaging"
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
	repo      domain.CampaignRepository
	publisher messaging.EventPublisher
}

func NewCampaignService(repo domain.CampaignRepository, publisher messaging.EventPublisher) *CampaignService {
	return &CampaignService{repo: repo, publisher: publisher}
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

	if err := s.repo.Save(ctx, campaign); err != nil {
		return nil, fmt.Errorf("save campaign: %w", err)
	}

	s.publishEvents(ctx, campaign.PullEvents())
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

	if err := s.repo.Update(ctx, campaign); err != nil {
		return nil, fmt.Errorf("update campaign: %w", err)
	}

	s.publishEvents(ctx, campaign.PullEvents())
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

func (s *CampaignService) publishEvents(ctx context.Context, events []ddd.DomainEvent) {
	for _, event := range events {
		data, _ := json.Marshal(event)
		s.publisher.Publish(ctx, "crm."+event.EventType(), data)
	}
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
