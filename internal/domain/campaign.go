package domain

import (
	"time"

	"github.com/crm-system-new/crm-shared/pkg/ddd"
)

type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusPaused    CampaignStatus = "paused"
	CampaignStatusCompleted CampaignStatus = "completed"
)

// Campaign is an aggregate root representing a marketing campaign.
type Campaign struct {
	ddd.AggregateRoot
	Name        string
	Description string
	Status      CampaignStatus
	Channel     string // email, sms, push, etc.
	SegmentID   string
	ScheduledAt *time.Time
	SentCount   int
	OpenRate    float64
	ClickRate   float64
}

// NewCampaign creates a new Campaign aggregate.
func NewCampaign(name, description, channel, segmentID string, scheduledAt *time.Time) (*Campaign, error) {
	if name == "" {
		return nil, ddd.ErrValidation{Field: "name", Message: "name is required"}
	}
	if channel == "" {
		return nil, ddd.ErrValidation{Field: "channel", Message: "channel is required"}
	}

	c := &Campaign{
		AggregateRoot: ddd.NewAggregateRoot(),
		Name:          name,
		Description:   description,
		Status:        CampaignStatusDraft,
		Channel:       channel,
		SegmentID:     segmentID,
		ScheduledAt:   scheduledAt,
	}

	return c, nil
}

// Launch activates the campaign for sending.
func (c *Campaign) Launch() error {
	if c.Status != CampaignStatusDraft && c.Status != CampaignStatusPaused {
		return ddd.ErrValidation{Field: "status", Message: "can only launch draft or paused campaigns"}
	}
	if c.SegmentID == "" {
		return ddd.ErrValidation{Field: "segment_id", Message: "cannot launch without a segment"}
	}

	c.Status = CampaignStatusActive
	c.IncrementVersion()

	c.RaiseEvent(CampaignLaunched{
		BaseEvent: ddd.NewBaseEvent("marketing.campaign.launched", c.ID),
		Name:      c.Name,
		Channel:   c.Channel,
		SegmentID: c.SegmentID,
	})

	return nil
}

// Pause temporarily stops a running campaign.
func (c *Campaign) Pause() error {
	if c.Status != CampaignStatusActive {
		return ddd.ErrValidation{Field: "status", Message: "can only pause active campaigns"}
	}

	c.Status = CampaignStatusPaused
	c.IncrementVersion()
	return nil
}

// Complete marks the campaign as finished.
func (c *Campaign) Complete() error {
	if c.Status != CampaignStatusActive {
		return ddd.ErrValidation{Field: "status", Message: "can only complete active campaigns"}
	}

	c.Status = CampaignStatusCompleted
	c.IncrementVersion()

	c.RaiseEvent(CampaignCompleted{
		BaseEvent: ddd.NewBaseEvent("marketing.campaign.completed", c.ID),
		SentCount: c.SentCount,
		OpenRate:  c.OpenRate,
		ClickRate: c.ClickRate,
	})

	return nil
}
