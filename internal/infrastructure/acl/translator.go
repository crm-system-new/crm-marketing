package acl

// AddSubscriberCommand is the local domain command produced by translating
// external Sales events into something the Marketing context understands.
type AddSubscriberCommand struct {
	Email     string
	FirstName string
	LastName  string
	Source    string
}

// TranslateLeadCreated converts a Sales LeadCreated event into a local
// AddSubscriberCommand.
func TranslateLeadCreated(event SalesLeadCreatedEvent) *AddSubscriberCommand {
	return &AddSubscriberCommand{
		Email:     event.Email,
		FirstName: event.FirstName,
		LastName:  event.LastName,
		Source:    event.Source,
	}
}

// TranslateCustomerCreated converts a Sales CustomerCreated event into a
// local AddSubscriberCommand.
func TranslateCustomerCreated(event SalesCustomerCreatedEvent) *AddSubscriberCommand {
	return &AddSubscriberCommand{
		Email:     event.Email,
		FirstName: event.Name,
		LastName:  "",
		Source:    "customer",
	}
}
