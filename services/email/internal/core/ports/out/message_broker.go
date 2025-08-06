package out

type MessageBrokerPort interface {
	ConsumeMessages(queueName string, handler func([]byte) error) error
	Close() error
}
