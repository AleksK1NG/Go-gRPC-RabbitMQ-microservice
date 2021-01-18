package main

import (
	"log"
	"os"

	"github.com/opentracing/opentracing-go"

	"github.com/AleksK1NG/email-microservice/config"
	"github.com/AleksK1NG/email-microservice/internal/server"
	"github.com/AleksK1NG/email-microservice/pkg/jaeger"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/AleksK1NG/email-microservice/pkg/mailer"
	"github.com/AleksK1NG/email-microservice/pkg/postgres"
	"github.com/AleksK1NG/email-microservice/pkg/rabbitmq"
)

func main() {
	log.Println("Starting server")

	configPath := config.GetConfigPath(os.Getenv("config"))
	cfg, err := config.GetConfig(configPath)
	if err != nil {
		log.Fatalf("Loading config: %v", err)
	}

	appLogger := logger.NewApiLogger(cfg)
	appLogger.InitLogger()
	appLogger.Infof(
		"AppVersion: %s, LogLevel: %s, Mode: %s, SSL: %v",
		cfg.Server.AppVersion,
		cfg.Logger.Level,
		cfg.Server.Mode,
		cfg.Server.SSL,
	)
	appLogger.Infof("Success parsed config: %#v", cfg.Server.AppVersion)

	amqpConn, err := rabbitmq.NewRabbitMQConn(cfg)
	if err != nil {
		appLogger.Fatal(err)
	}
	defer amqpConn.Close()

	psqlDB, err := postgres.NewPsqlDB(cfg)
	if err != nil {
		appLogger.Fatalf("Postgresql init: %s", err)
	}
	defer psqlDB.Close()

	appLogger.Infof("PostgreSQL connected: %#v", psqlDB.Stats())

	tracer, closer, err := jaeger.InitJaeger(cfg)
	if err != nil {
		appLogger.Fatal("cannot create tracer", err)
	}
	appLogger.Info("Jaeger connected")

	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
	appLogger.Info("Opentracing connected")

	mailDialer := mailer.NewMailDialer(cfg)
	appLogger.Info("Mail dialer connected")

	s := server.NewEmailsServer(amqpConn, appLogger, cfg, mailDialer, psqlDB)

	appLogger.Fatal(s.Run())
}
