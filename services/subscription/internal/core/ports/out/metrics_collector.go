package out

type MetricsCollector interface {
	IncrementHTTPRequests(method, path string, statusCode int)
	RecordHTTPDuration(method, path string, duration float64)

	IncrementSubscriptionCreated()
	IncrementSubscriptionConfirmed()
	IncrementSubscriptionUnsubscribed()
	IncrementSubscriptionErrors()

	IncrementRabbitMQPublished()
	IncrementRabbitMQPublishErrors()

	RecordDatabaseOperation(operation string, duration float64)
	IncrementDatabaseErrors()

	IncrementGRPCRequests(service, method string)
	IncrementGRPCErrors(service, method string)
	RecordGRPCDuration(service, method string, duration float64)
}
