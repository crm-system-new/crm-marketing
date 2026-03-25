package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/crm-system-new/crm-marketing/internal/domain"
	"github.com/crm-system-new/crm-shared/pkg/messaging"
	natssub "github.com/crm-system-new/crm-shared/pkg/messaging/nats"
)

// SalesEventSubscriber consumes events from the Sales bounded context
// and auto-creates marketing subscribers from leads and customers.
type SalesEventSubscriber struct {
	subscriber     *natssub.Subscriber
	subscriberRepo domain.SubscriberRepository
}

// NewSalesEventSubscriber creates a subscriber that listens to sales events.
func NewSalesEventSubscriber(natsURL string, subscriberRepo domain.SubscriberRepository) (*SalesEventSubscriber, error) {
	sub, err := natssub.NewSubscriber(natsURL, "marketing-service")
	if err != nil {
		return nil, err
	}

	s := &SalesEventSubscriber{
		subscriber:     sub,
		subscriberRepo: subscriberRepo,
	}

	if err := s.setup(); err != nil {
		sub.Close()
		return nil, err
	}

	return s, nil
}

func (s *SalesEventSubscriber) setup() error {
	if err := s.subscriber.Subscribe("crm.sales.lead.created", s.handleLeadCreated); err != nil {
		return err
	}
	if err := s.subscriber.Subscribe("crm.sales.customer.created", s.handleCustomerCreated); err != nil {
		return err
	}
	return nil
}

func (s *SalesEventSubscriber) handleLeadCreated(_ context.Context, _ string, data []byte) error {
	var event struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	if event.Email == "" {
		return nil
	}

	return s.ensureSubscriber(event.Email, event.FirstName, event.LastName)
}

func (s *SalesEventSubscriber) handleCustomerCreated(_ context.Context, _ string, data []byte) error {
	var event struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	if event.Email == "" {
		return nil
	}

	return s.ensureSubscriber(event.Email, event.Name, "")
}

func (s *SalesEventSubscriber) ensureSubscriber(email, firstName, lastName string) error {
	ctx := context.Background()

	// Check if subscriber already exists
	_, err := s.subscriberRepo.GetByEmail(ctx, email)
	if err == nil {
		// Already exists, skip
		return nil
	}

	subscriber, err := domain.NewSubscriber(email, firstName, lastName)
	if err != nil {
		log.Printf("WARN: failed to create subscriber from sales event: %v", err)
		return nil
	}

	// Discard events since these are internally triggered
	subscriber.PullEvents()

	if err := s.subscriberRepo.Save(ctx, subscriber); err != nil {
		log.Printf("WARN: failed to save subscriber from sales event: %v", err)
		return nil
	}

	log.Printf("Auto-created subscriber %s from sales event", email)
	return nil
}

// Expose the underlying subscriber's Close for use by main.
func (s *SalesEventSubscriber) Close() error {
	return s.subscriber.Close()
}

// Ensure SalesEventSubscriber satisfies an implicit closeable contract.
var _ interface{ Close() error } = (*SalesEventSubscriber)(nil)

// Ensure the underlying natssub.Subscriber satisfies EventSubscriber.
var _ messaging.EventSubscriber = (*natssub.Subscriber)(nil)
