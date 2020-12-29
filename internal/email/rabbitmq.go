package email

import "github.com/streadway/amqp"

// Emails publisher interface
type EmailsPublisher interface {
	Publish(body []byte, contentType string) error
	CloseChan()
	SetupExchangeAndQueue(exchange, queueName, bindingKey, consumerTag string) error
}

// Emails consumer interface
type EmailsConsumer interface {
	CreateChannel(exchangeName, queueName, bindingKey, consumerTag string) (*amqp.Channel, error)
	StartConsumer(workerPoolSize int, exchange, queueName, bindingKey, consumerTag string) error
}
