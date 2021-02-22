package rabbitmq

import (
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/streadway/amqp"

	"github.com/AleksK1NG/email-microservice/config"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/AleksK1NG/email-microservice/pkg/rabbitmq"
)

var (
	publishedMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "emails_published_rabbitmq_messages_total",
		Help: "The total number of published RabbitMQ messages",
	})
)

// EmailsPublisher Emails rabbitmq publisher
type EmailsPublisher struct {
	amqpChan *amqp.Channel
	cfg      *config.Config
	logger   logger.Logger
}

// NewEmailsPublisher Emails rabbitmq publisher constructor
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

// SetupExchangeAndQueue create exchange and queue
func (p *EmailsPublisher) SetupExchangeAndQueue(exchange, queueName, bindingKey, consumerTag string) error {
	p.logger.Infof("Declaring exchange: %s", exchange)
	err := p.amqpChan.ExchangeDeclare(
		exchange,
		exchangeKind,
		exchangeDurable,
		exchangeAutoDelete,
		exchangeInternal,
		exchangeNoWait,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "Error ch.ExchangeDeclare")
	}

	queue, err := p.amqpChan.QueueDeclare(
		queueName,
		queueDurable,
		queueAutoDelete,
		queueExclusive,
		queueNoWait,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "Error ch.QueueDeclare")
	}

	p.logger.Infof("Declared queue, binding it to exchange: Queue: %v, messageCount: %v, "+
		"consumerCount: %v, exchange: %v, exchange: %v, bindingKey: %v",
		queue.Name,
		queue.Messages,
		queue.Consumers,
		exchange,
		bindingKey,
	)

	err = p.amqpChan.QueueBind(
		queue.Name,
		bindingKey,
		exchange,
		queueNoWait,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "Error ch.QueueBind")
	}

	p.logger.Infof("Queue bound to exchange, starting to consume from queue, consumerTag: %v", consumerTag)
	return nil
}

// CloseChan Close messages chan
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
		publishMandatory,
		publishImmediate,
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

	publishedMessages.Inc()
	return nil
}
