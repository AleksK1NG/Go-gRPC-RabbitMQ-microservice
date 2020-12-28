package rabbitmq

import (
	"github.com/AleksK1NG/email-microservice/config"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/AleksK1NG/email-microservice/pkg/rabbitmq"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"time"
)

// Emails rabbitmq publisher
type EmailsPublisher struct {
	amqpChan *amqp.Channel
	cfg      *config.Config
	logger   logger.Logger
}

// Emails rabbitmq publisher constructor
func NewEmailsPublisher(cfg *config.Config, logger logger.Logger) (*EmailsPublisher, error) {
	mqConn, err := rabbitmq.NewRabbitMQConn(cfg)
	if err != nil {
		return nil, err
	}
	amqpChan, err := mqConn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "p.amqpConn.Channel")
	}

	return &EmailsPublisher{cfg: cfg, logger: logger, amqpChan: amqpChan}, nil
}

func (e *EmailsPublisher) SetupExchangeAndQueue(exchange, queueName, bindingKey, consumerTag string) error {
	e.logger.Infof("Declaring exchange: %s", exchange)
	err := e.amqpChan.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "Error ch.ExchangeDeclare")
	}

	queue, err := e.amqpChan.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "Error ch.QueueDeclare")
	}

	e.logger.Infof("Declared queue, binding it to exchange: Queue: %v, messageCount: %v, "+
		"consumerCount: %v, exchange: %v, exchange: %v, bindingKey: %v",
		queue.Name,
		queue.Messages,
		queue.Consumers,
		exchange,
		bindingKey,
	)

	err = e.amqpChan.QueueBind(
		queue.Name,
		bindingKey,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "Error ch.QueueBind")
	}

	e.logger.Infof("Queue bound to exchange, starting to consume from queue, consumerTag: %v", consumerTag)
	return nil
}

// Close messages chan
func (p *EmailsPublisher) CloseChan() {
	if err := p.amqpChan.Close(); err != nil {
		p.logger.Errorf("EmailsPublisher CloseChan: %v", err)
	}
}

// Publish message
func (p *EmailsPublisher) Publish(body []byte, contentType string) error {

	p.logger.Infof("Publishing message Exchange: %s, RoutingKey: %s", p.cfg.RabbitMQ.Exchange, p.cfg.RabbitMQ.RoutingKey)

	if err := p.amqpChan.Publish(
		p.cfg.RabbitMQ.Exchange,
		p.cfg.RabbitMQ.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  contentType,
			DeliveryMode: amqp.Persistent,
			MessageId:    uuid.New().String(),
			Timestamp:    time.Now(),
			Body:         body,
		},
	); err != nil {
		return errors.Wrap(err, "ch.Publish")
	}

	return nil
}
