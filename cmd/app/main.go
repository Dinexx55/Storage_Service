package main

import (
	"StorageService/internal/config"
	"StorageService/internal/migration"
	"StorageService/internal/repository/postgres"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"log"
	"os"
	"time"
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

	migrator := migration.NewMigratory()

	repository, err := initDBWithRetry(cfg, migrator, logger)
	if err != nil {
		logger.With(
			zap.String("place", "main"),
			zap.Error(err),
		).Panic("Failed to establish database connection")
	}
	defer repository.Close()

	rabbitConnection, err := initRabbitMQConnection(cfg)

	if err != nil {
		logger.With(
			zap.String("place", "main"),
			zap.Error(err),
		).Panic("Failed to establish RabbitMQ rabbitConnection")
	}

	channel, err := initRabbitChannel(rabbitConnection)

	if err != nil {
		logger.With(
			zap.String("place", "main"),
			zap.Error(err),
		).Panic("Failed to init RabbitMQ channel")
	}

	queue, err := declareRabbitQueue(channel)

	if err != nil {
		logger.With(
			zap.String("place", "main"),
			zap.Error(err),
		).Panic("Failed to init RabbitMQ queue")
	}

	msgs, err := channel.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)

	if err != nil {
		logger.With(
			zap.String("place", "main"),
			zap.Error(err),
		).Panic("Failed to register a consumer")
	}

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func declareRabbitQueue(channel *amqp.Channel) (amqp.Queue, error) {
	queue, err := channel.QueueDeclare(
		"CreateQueue", // name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	return queue, err
}

func initRabbitChannel(connection *amqp.Connection) (*amqp.Channel, error) {
	channel, err := connection.Channel()

	return channel, err
}

func initRabbitMQConnection(cfg *config.Configurator) (*amqp.Connection, error) {
	mqConfig := cfg.GetRabbitMQConfig()

	conn, err := amqp.Dial(cfg.GetAMQPConnectionURL(mqConfig))

	return conn, err
}

func initLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()

	if os.Getenv("APP_ENV") == "development" {
		logger, err = zap.NewDevelopment()
	}

	return logger, err
}

func initDBWithRetry(cfg *config.Configurator, migrator *migration.Migratory, logger *zap.Logger) (repo *postgres.Repository, err error) {
	logger.Info("Getting cfg for postgres")

	dbCfg, err := cfg.DBConfig()
	if err != nil {
		return nil, err
	}

	logger.Info("Got db config")

	var db *sqlx.DB
	for i := 0; i < dbCfg.ReconnRetry; i++ {

		db, err = postgres.ConnectToPostgresDB(dbCfg)

		if err == nil {

			logger.Info("Db migration")

			if err = migrator.Migrate(db); err != nil {
				return nil, fmt.Errorf("migration failure: %w", err)
			}

			repo = postgres.NewPostgresRepository(db)

			logger.Info("Migrations done")

			return repo, nil
		}

		logger.With(
			zap.String("place", "main"),
			zap.Error(err),
		).Error("Failed to connect to db. Retrying")

		time.Sleep(dbCfg.TimeWaitPerTry)
	}
	logger.Info("Successfully connected to postgres")
	return repo, err
}
