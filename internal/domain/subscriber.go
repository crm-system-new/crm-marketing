package domain

import (
	"github.com/crm-system-new/crm-shared/pkg/ddd"
)

type SubscriberStatus string

const (
	SubscriberStatusActive       SubscriberStatus = "active"
	SubscriberStatusUnsubscribed SubscriberStatus = "unsubscribed"
	SubscriberStatusBounced      SubscriberStatus = "bounced"
)

// Subscriber is an aggregate root representing a marketing subscriber.
type Subscriber struct {
	ddd.AggregateRoot
	Email       string
	FirstName   string
	LastName    string
	Status      SubscriberStatus
	Preferences string // JSON map of preferences
}

// NewSubscriber creates a new Subscriber aggregate.
func NewSubscriber(email, firstName, lastName string) (*Subscriber, error) {
	if email == "" {
		return nil, ddd.ErrValidation{Field: "email", Message: "email is required"}
	}

	s := &Subscriber{
		AggregateRoot: ddd.NewAggregateRoot(),
		Email:         email,
		FirstName:     firstName,
		LastName:      lastName,
		Status:        SubscriberStatusActive,
		Preferences:   "{}",
	}

	s.RaiseEvent(SubscriberAdded{
		BaseEvent: ddd.NewBaseEvent("marketing.subscriber.added", s.ID),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	})

	return s, nil
}

// Unsubscribe marks the subscriber as unsubscribed.
func (s *Subscriber) Unsubscribe() error {
	if s.Status == SubscriberStatusUnsubscribed {
		return ddd.ErrValidation{Field: "status", Message: "subscriber is already unsubscribed"}
	}

	s.Status = SubscriberStatusUnsubscribed
	s.IncrementVersion()

	s.RaiseEvent(SubscriberUnsubscribed{
		BaseEvent: ddd.NewBaseEvent("marketing.subscriber.unsubscribed", s.ID),
		Email:     s.Email,
	})

	return nil
}

// UpdatePreferences updates the subscriber's preferences JSON.
func (s *Subscriber) UpdatePreferences(preferences string) error {
	s.Preferences = preferences
	s.IncrementVersion()
	return nil
}
