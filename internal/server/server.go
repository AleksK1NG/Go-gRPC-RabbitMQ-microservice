package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"gopkg.in/gomail.v2"

	"github.com/AleksK1NG/email-microservice/config"
	mailGrpc "github.com/AleksK1NG/email-microservice/internal/email/delivery/grpc"
	"github.com/AleksK1NG/email-microservice/internal/email/delivery/rabbitmq"
	"github.com/AleksK1NG/email-microservice/internal/email/mailer"
	emailService "github.com/AleksK1NG/email-microservice/internal/email/proto"
	"github.com/AleksK1NG/email-microservice/internal/email/repository"
	"github.com/AleksK1NG/email-microservice/internal/email/usecase"
	"github.com/AleksK1NG/email-microservice/internal/interceptors"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/AleksK1NG/email-microservice/pkg/metrics"
)

// Server
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
	metric, err := metrics.CreateMetrics(s.cfg.Metrics.URL, s.cfg.Metrics.ServiceName)
	if err != nil {
		s.logger.Errorf("CreateMetrics Error: %s", err)
	}
	s.logger.Info(
		"Metrics available URL: %s, ServiceName: %s",
		s.cfg.Metrics.URL,
		s.cfg.Metrics.ServiceName,
	)

	emailsPublisher, err := rabbitmq.NewEmailsPublisher(s.cfg, s.logger)
	if err != nil {
		return err
	}
	defer emailsPublisher.CloseChan()
	s.logger.Info("Emails Publisher initialized")

	im := interceptors.NewInterceptorManager(s.logger, s.cfg, metric)

	emailRepository := repository.NewEmailsRepository(s.db)
	mailDialer := mailer.NewMailer(s.mailDialer)
	emailUseCase := usecase.NewEmailUseCase(emailRepository, s.logger, mailDialer, s.cfg, emailsPublisher)
	emailsAmqpConsumer := rabbitmq.NewImagesConsumer(s.amqpConn, s.logger, emailUseCase)

	ctx, cancel := context.WithCancel(context.Background())

	router := echo.New()
	router.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	go func() {
		if err := router.Start(s.cfg.Metrics.URL); err != nil {
			s.logger.Errorf("router.Start metrics: %v", err)
			cancel()
		}
	}()

	go func() {
		err := emailsAmqpConsumer.StartConsumer(
			s.cfg.RabbitMQ.WorkerPoolSize,
			s.cfg.RabbitMQ.Exchange,
			s.cfg.RabbitMQ.Queue,
			s.cfg.RabbitMQ.RoutingKey,
			s.cfg.RabbitMQ.ConsumerTag,
		)
		if err != nil {
			s.logger.Errorf("StartConsumer: %v", err)
			cancel()
		}
	}()

	l, err := net.Listen("tcp", s.cfg.Server.Port)
	if err != nil {
		return err
	}
	defer l.Close()

	server := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: s.cfg.Server.MaxConnectionIdle * time.Minute,
		Timeout:           s.cfg.Server.Timeout * time.Second,
		MaxConnectionAge:  s.cfg.Server.MaxConnectionAge * time.Minute,
		Time:              s.cfg.Server.Timeout * time.Minute,
	}),
		grpc.UnaryInterceptor(im.Logger),
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpcrecovery.UnaryServerInterceptor(),
		),
	)

	emailGrpcMicroservice := mailGrpc.NewEmailMicroservice(emailUseCase, s.logger, s.cfg)
	emailService.RegisterEmailServiceServer(server, emailGrpcMicroservice)
	grpc_prometheus.Register(server)
	s.logger.Info("Emails Service initialized")

	if s.cfg.Server.Mode != "Production" {
		reflection.Register(server)
	}

	go func() {
		s.logger.Infof("Server is listening on port: %v", s.cfg.Server.Port)
		s.logger.Fatal(server.Serve(l))
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case v := <-quit:
		s.logger.Errorf("signal.Notify: %v", v)
	case done := <-ctx.Done():
		s.logger.Errorf("ctx.Done: %v", done)
	}

	if err := router.Shutdown(ctx); err != nil {
		s.logger.Errorf("Metrics router.Shutdown: %v", err)
	}
	server.GracefulStop()
	s.logger.Info("Server Exited Properly")

	return nil
}
