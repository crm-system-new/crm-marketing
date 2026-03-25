package messaging

import (
	natspub "github.com/crm-system-new/crm-shared/pkg/messaging/nats"
)

// NewMarketingPublisher creates a NATS publisher and ensures the MARKETING stream exists.
func NewMarketingPublisher(natsURL string) (*natspub.Publisher, error) {
	pub, err := natspub.NewPublisher(natsURL)
	if err != nil {
		return nil, err
	}

	if err := pub.EnsureStream("MARKETING", []string{"crm.marketing.>"}); err != nil {
		pub.Close()
		return nil, err
	}

	return pub, nil
}
