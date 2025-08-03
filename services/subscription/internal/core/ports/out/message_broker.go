package out

import "subscription/internal/core/domain"

// MessageBrokerPort defines the interface for message broker operations
type MessageBrokerPort interface {
	// PublishSubscriptionCreated publishes subscription created event
	PublishSubscriptionCreated(event domain.SubscriptionEvent) error

	// PublishSubscriptionConfirmed publishes subscription confirmed event
	PublishSubscriptionConfirmed(event domain.ConfirmedEvent) error

	// PublishSubscriptionUnsubscribed publishes subscription unsubscribed event
	PublishSubscriptionUnsubscribed(event domain.UnsubscribedEvent) error

	// Close closes the connection to the message broker
	Close() error
}
