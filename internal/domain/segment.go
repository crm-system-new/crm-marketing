package domain

import (
	"github.com/crm-system-new/crm-shared/pkg/ddd"
)

// Segment is an aggregate root representing a subscriber segment for targeting.
type Segment struct {
	ddd.AggregateRoot
	Name            string
	Criteria        string // JSON criteria for segment filtering
	SubscriberCount int
}

// NewSegment creates a new Segment aggregate.
func NewSegment(name, criteria string) (*Segment, error) {
	if name == "" {
		return nil, ddd.ErrValidation{Field: "name", Message: "name is required"}
	}

	s := &Segment{
		AggregateRoot: ddd.NewAggregateRoot(),
		Name:          name,
		Criteria:      criteria,
	}

	return s, nil
}

// UpdateDetails updates the segment's name and criteria.
func (s *Segment) UpdateDetails(name, criteria string) error {
	if name == "" {
		return ddd.ErrValidation{Field: "name", Message: "name is required"}
	}
	s.Name = name
	s.Criteria = criteria
	s.IncrementVersion()
	return nil
}
