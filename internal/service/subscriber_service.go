package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/crm-system-new/crm-marketing/internal/domain"
	"github.com/crm-system-new/crm-shared/pkg/ddd"
	"github.com/crm-system-new/crm-shared/pkg/messaging"
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
	repo      domain.SubscriberRepository
	publisher messaging.EventPublisher
}

func NewSubscriberService(repo domain.SubscriberRepository, publisher messaging.EventPublisher) *SubscriberService {
	return &SubscriberService{repo: repo, publisher: publisher}
}

func (s *SubscriberService) Subscribe(ctx context.Context, req SubscribeRequest) (*SubscriberResponse, error) {
	subscriber, err := domain.NewSubscriber(req.Email, req.FirstName, req.LastName)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, subscriber); err != nil {
		return nil, fmt.Errorf("save subscriber: %w", err)
	}

	s.publishEvents(ctx, subscriber.PullEvents())
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

	if err := s.repo.Update(ctx, subscriber); err != nil {
		return nil, fmt.Errorf("update subscriber: %w", err)
	}

	s.publishEvents(ctx, subscriber.PullEvents())
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

func (s *SubscriberService) publishEvents(ctx context.Context, events []ddd.DomainEvent) {
	for _, event := range events {
		data, _ := json.Marshal(event)
		s.publisher.Publish(ctx, "crm."+event.EventType(), data)
	}
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
