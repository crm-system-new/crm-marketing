package domain

import "github.com/crm-system-new/crm-shared/pkg/ddd"

// Campaign events

type CampaignLaunched struct {
	ddd.BaseEvent
	Name      string `json:"name"`
	Channel   string `json:"channel"`
	SegmentID string `json:"segment_id"`
}

type CampaignCompleted struct {
	ddd.BaseEvent
	SentCount int     `json:"sent_count"`
	OpenRate  float64 `json:"open_rate"`
	ClickRate float64 `json:"click_rate"`
}

// Subscriber events

type SubscriberAdded struct {
	ddd.BaseEvent
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type SubscriberUnsubscribed struct {
	ddd.BaseEvent
	Email string `json:"email"`
}
