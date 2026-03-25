package domain

import "context"

// CampaignFilter defines filters for listing campaigns.
type CampaignFilter struct {
	Status    *CampaignStatus
	Channel   *string
	SegmentID *string
	Limit     int
	Offset    int
}

// SegmentFilter defines filters for listing segments.
type SegmentFilter struct {
	Limit  int
	Offset int
}

// SubscriberFilter defines filters for listing subscribers.
type SubscriberFilter struct {
	Status *SubscriberStatus
	Email  *string
	Limit  int
	Offset int
}

type CampaignRepository interface {
	GetByID(ctx context.Context, id string) (*Campaign, error)
	Save(ctx context.Context, campaign *Campaign) error
	Update(ctx context.Context, campaign *Campaign) error
	List(ctx context.Context, filter CampaignFilter) ([]*Campaign, int, error)
}

type SegmentRepository interface {
	GetByID(ctx context.Context, id string) (*Segment, error)
	Save(ctx context.Context, segment *Segment) error
	Update(ctx context.Context, segment *Segment) error
	List(ctx context.Context, filter SegmentFilter) ([]*Segment, int, error)
}

type SubscriberRepository interface {
	GetByID(ctx context.Context, id string) (*Subscriber, error)
	GetByEmail(ctx context.Context, email string) (*Subscriber, error)
	Save(ctx context.Context, subscriber *Subscriber) error
	Update(ctx context.Context, subscriber *Subscriber) error
	List(ctx context.Context, filter SubscriberFilter) ([]*Subscriber, int, error)
}
