package server

import (
	"github.com/AleksK1NG/email-microservice/config"
	"github.com/AleksK1NG/email-microservice/internal/email/delivery/rabbitmq"
	"github.com/AleksK1NG/email-microservice/internal/email/mailer"
	"github.com/AleksK1NG/email-microservice/internal/email/repository"
	"github.com/AleksK1NG/email-microservice/internal/email/usecase"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/streadway/amqp"
	"gopkg.in/gomail.v2"
	"os"
	"os/signal"
	"syscall"
)

// Images service
type Server struct {
	db         *sqlx.DB
	mailDialer *gomail.Dialer
	amqpConn   *amqp.Connection
	logger     logger.Logger
	cfg        *config.Config
}

// Server constructor
func NewEmailsServer(amqpConn *amqp.Connection, logger logger.Logger, cfg *config.Config, mailDialer *gomail.Dialer, db *sqlx.DB) *Server {
	return &Server{amqpConn: amqpConn, logger: logger, cfg: cfg, mailDialer: mailDialer, db: db}
}

// Run server
func (s *Server) Run() error {
	emailRepository := repository.NewEmailsRepository(s.db)
	mailDialer := mailer.NewMailer(s.cfg, s.mailDialer)
	emailUseCase := usecase.NewEmailUseCase(emailRepository, s.logger, mailDialer)
	emailsAmqpConsumer := rabbitmq.NewImagesConsumer(s.amqpConn, s.logger, emailUseCase)

	go func() {
		err := emailsAmqpConsumer.StartConsumer(
			s.cfg.RabbitMQ.WorkerPoolSize,
			s.cfg.RabbitMQ.Exchange,
			s.cfg.RabbitMQ.Queue, "",
			s.cfg.RabbitMQ.ConsumerTag,
		)
		if err != nil {
			s.logger.Fatalf("StartConsumer: %v", err)
		}
	}()

	go func() {
		router := echo.New()
		router.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
		if err := router.Start(s.cfg.Metrics.URL); err != nil {
			s.logger.Fatalf("router.Start metrics: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	s.logger.Info("Server Exited Properly")
	return nil
}
