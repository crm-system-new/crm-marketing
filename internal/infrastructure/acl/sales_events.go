package acl

// SalesLeadCreatedEvent is an external event DTO that mirrors the Sales
// bounded context's LeadCreated event schema. We do NOT import crm-sales;
// instead we maintain our own copy to form an Anti-Corruption Layer.
type SalesLeadCreatedEvent struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Source    string `json:"source"`
}

// SalesCustomerCreatedEvent is an external event DTO that mirrors the Sales
// bounded context's CustomerCreated event schema.
type SalesCustomerCreatedEvent struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	OwnerID   string `json:"owner_id"`
}
