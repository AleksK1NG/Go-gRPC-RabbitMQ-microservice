package rabbitmq

import (
	"context"
	"github.com/AleksK1NG/email-microservice/internal/email"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/streadway/amqp"
	"sync"
)

var (
	incomingMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "emails_incoming_rabbitmq_messages_total",
		Help: "The total number of incoming RabbitMQ messages",
	})
	successMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "emails_success_incoming_rabbitmq_messages_total",
		Help: "The total number of success incoming success RabbitMQ messages",
	})
	errorMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "emails_error_incoming_rabbitmq_message_total",
		Help: "The total number of error incoming success RabbitMQ messages",
	})
)

// Images Rabbitmq consumer
type EmailsConsumer struct {
	amqpConn *amqp.Connection
	logger   logger.Logger
	emailUC  email.EmailsUseCase
}

// Images Consumer constructor
func NewImagesConsumer(amqpConn *amqp.Connection, logger logger.Logger, emailUC email.EmailsUseCase) *EmailsConsumer {
	return &EmailsConsumer{amqpConn: amqpConn, logger: logger, emailUC: emailUC}
}

// Consume messages
func (c *EmailsConsumer) CreateChannel(exchangeName, queueName, bindingKey, consumerTag string) (*amqp.Channel, error) {
	ch, err := c.amqpConn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "Error amqpConn.Channel")
	}

	c.logger.Infof("Declaring exchange: %s", exchangeName)
	err = ch.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error ch.ExchangeDeclare")
	}

	queue, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error ch.QueueDeclare")
	}

	c.logger.Infof("Declared queue, binding it to exchange: Queue: %v, messageCount: %v, "+
		"consumerCount: %v, exchange: %v, exchange: %v, bindingKey: %v",
		queue.Name,
		queue.Messages,
		queue.Consumers,
		exchangeName,
		bindingKey,
	)

	err = ch.QueueBind(
		queue.Name,
		bindingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error ch.QueueBind")
	}

	c.logger.Infof("Queue bound to exchange, starting to consume from queue, consumerTag: %v", consumerTag)

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error  ch.Qos")
	}

	return ch, nil
}

func (c *EmailsConsumer) worker(ctx context.Context, messages <-chan amqp.Delivery, wg *sync.WaitGroup) {
	defer wg.Done()

	for delivery := range messages {
		span, ctx := opentracing.StartSpanFromContext(ctx, "EmailsConsumer.worker")

		c.logger.Infof("processDeliveries deliveryTag% v", delivery.DeliveryTag)

		incomingMessages.Inc()

		err := c.emailUC.SendEmail(ctx, delivery)
		if err != nil {
			if err := delivery.Reject(false); err != nil {
				c.logger.Errorf("Err delivery.Reject: %v", err)
			}
			errorMessages.Inc()
			c.logger.Errorf("Failed to process delivery: %v", err)
			span.Finish()
		} else {
			successMessages.Inc()
			err = delivery.Ack(false)
			if err != nil {
				c.logger.Errorf("Failed to acknowledge delivery: %v", err)
			}
			span.Finish()
		}
	}

	c.logger.Info("Deliveries channel closed")
}

// Start new rabbitmq consumer
func (c *EmailsConsumer) StartConsumer(workerPoolSize int, exchange, queueName, bindingKey, consumerTag string) error {
	ch, err := c.CreateChannel(exchange, queueName, bindingKey, consumerTag)
	if err != nil {
		return err
	}
	defer ch.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deliveries, err := ch.Consume(
		queueName,
		consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)

	wg := &sync.WaitGroup{}
	for i := 0; i < workerPoolSize; i++ {
		wg.Add(1)
		go c.worker(ctx, deliveries, wg)
	}

	wg.Wait()
	return nil
}
