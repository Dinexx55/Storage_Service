package main

import (
	"StorageService/internal/config"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"log"
	"os"
)

func main() {
	cfg, err := config.NewConfiguration()

	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	logger, err := initLogger()

	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	if isRelease := cfg.GetEnvironment(logger) == config.Release; isRelease {
		logger.Info("Got application environment. Running in Release")
	} else {
		logger.Info("Got application environment. Running in Development")
	}

	//rabbitChannel, err := initRabbitMQConnection(cfg, logger)
}

func initRabbitMQConnection(cfg *config.Configurator, logger *zap.Logger) (*amqp.Channel, error) {
	mqConfig := cfg.GetRabbitMQConfig()

	conn, err := amqp.Dial(cfg.GetAMQPConnectionURL(mqConfig))
	if err != nil {
		logger.With(
			zap.String("place", "initRabbitMQConnection"),
			zap.Error(err),
		).Error("Failed to establish RabbitMQ connection")
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		logger.With(
			zap.String("place", "initRabbitMQConnection"),
			zap.Error(err),
		).Error("Failed to open RabbitMQ channel")
		return nil, err
	}

	return channel, nil
}

func initLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if os.Getenv("APP_ENV") == "development" {
		logger, err = zap.NewDevelopment()
	}
	return logger, err
}
