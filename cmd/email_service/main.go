package main

import (
	"github.com/AleksK1NG/email-microservice/config"
	"github.com/AleksK1NG/email-microservice/internal/server"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/AleksK1NG/email-microservice/pkg/rabbitmq"
	"log"
	"os"
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

	s := server.NewEmailsServer(amqpConn, appLogger, cfg)

	if err := s.Run(); err != nil {
		appLogger.Fatal(err)
	}

}
