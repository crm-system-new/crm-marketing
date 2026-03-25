package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jackc/pgx/v5"

	"github.com/crm-system-new/crm-marketing/internal/domain"
	"github.com/crm-system-new/crm-marketing/internal/infrastructure/acl"
	"github.com/crm-system-new/crm-marketing/internal/infrastructure/postgres"
	"github.com/crm-system-new/crm-shared/pkg/idempotency"
	"github.com/crm-system-new/crm-shared/pkg/messaging"
	natssub "github.com/crm-system-new/crm-shared/pkg/messaging/nats"
)

// SalesEventSubscriber consumes events from the Sales bounded context
// and auto-creates marketing subscribers from leads and customers.
// It uses the ACL to translate external events and the idempotency store
// to prevent duplicate processing.
type SalesEventSubscriber struct {
	subscriber       *natssub.Subscriber
	subscriberRepo   *postgres.SubscriberRepository
	idempotencyStore *idempotency.Store
}

// NewSalesEventSubscriber creates a subscriber that listens to sales events.
func NewSalesEventSubscriber(natsURL string, subscriberRepo *postgres.SubscriberRepository, idempotencyStore *idempotency.Store) (*SalesEventSubscriber, error) {
	sub, err := natssub.NewSubscriber(natsURL, "marketing-service")
	if err != nil {
		return nil, err
	}

	s := &SalesEventSubscriber{
		subscriber:       sub,
		subscriberRepo:   subscriberRepo,
		idempotencyStore: idempotencyStore,
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
	var event acl.SalesLeadCreatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	if event.Email == "" {
		return nil
	}

	cmd := acl.TranslateLeadCreated(event)

	return s.idempotencyStore.ProcessOnce(context.Background(), event.EventID, func(ctx context.Context, tx pgx.Tx) error {
		return s.ensureSubscriberInTx(ctx, tx, cmd)
	})
}

func (s *SalesEventSubscriber) handleCustomerCreated(_ context.Context, _ string, data []byte) error {
	var event acl.SalesCustomerCreatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	if event.Email == "" {
		return nil
	}

	cmd := acl.TranslateCustomerCreated(event)

	return s.idempotencyStore.ProcessOnce(context.Background(), event.EventID, func(ctx context.Context, tx pgx.Tx) error {
		return s.ensureSubscriberInTx(ctx, tx, cmd)
	})
}

func (s *SalesEventSubscriber) ensureSubscriberInTx(ctx context.Context, tx pgx.Tx, cmd *acl.AddSubscriberCommand) error {
	// Check if subscriber already exists
	_, err := s.subscriberRepo.GetByEmail(ctx, cmd.Email)
	if err == nil {
		// Already exists, skip
		return nil
	}

	subscriber, err := domain.NewSubscriber(cmd.Email, cmd.FirstName, cmd.LastName)
	if err != nil {
		log.Printf("WARN: failed to create subscriber from sales event: %v", err)
		return nil
	}

	// Discard events since these are internally triggered
	subscriber.PullEvents()

	if err := s.subscriberRepo.SaveInTx(ctx, tx, subscriber); err != nil {
		log.Printf("WARN: failed to save subscriber from sales event: %v", err)
		return nil
	}

	log.Printf("Auto-created subscriber %s from sales event", cmd.Email)
	return nil
}

// Close shuts down the underlying NATS subscriber.
func (s *SalesEventSubscriber) Close() error {
	return s.subscriber.Close()
}

// Ensure SalesEventSubscriber satisfies an implicit closeable contract.
var _ interface{ Close() error } = (*SalesEventSubscriber)(nil)

// Ensure the underlying natssub.Subscriber satisfies EventSubscriber.
var _ messaging.EventSubscriber = (*natssub.Subscriber)(nil)
