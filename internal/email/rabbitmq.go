package email

// Emails publisher interface
type EmailsPublisher interface {
	Publish(body []byte, contentType string) error
}

// Emails consumer interface
type EmailsConsumer interface {
	StartConsumer(workerPoolSize int, exchange, queueName, bindingKey, consumerTag string) error
}
