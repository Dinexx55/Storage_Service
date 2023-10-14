package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
)

type RabbitMQConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

type Configurator struct {
}

func NewConfiguration() (*Configurator, error) {
	godotenv.Load("../../.env")
	switch os.Getenv("CURRENT_ENV") {
	case "local":

		viper.AddConfigPath("../../configs")
		viper.SetConfigName("config")
	default:

		viper.AddConfigPath("../../configs")
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read conf file: %w", err)
	}

	c := &Configurator{}

	return c, nil
}

type AppEnvironment string

const (
	Release             AppEnvironment = "release"
	Development         AppEnvironment = "development"
	DefaultEnv          AppEnvironment = Development
	EnvironmentVariable                = "APP_ENV"
)

func (cfg *Configurator) GetEnvironment(logger *zap.Logger) AppEnvironment {
	logger.With(
		zap.String("place", "GetEnvironment"),
	).Info("Reading GetEnvironment")

	env := os.Getenv(EnvironmentVariable)
	if env == "" {
		env = string(DefaultEnv)
	}

	logger.Info("Running in " + env)
	return AppEnvironment(env)
}

func (cfg *Configurator) GetRabbitMQConfig() *RabbitMQConfig {
	return &RabbitMQConfig{
		Password: viper.GetString("rabbit.password"),
		Username: viper.GetString("rabbit.username"),
		Port:     viper.GetString("rabbit.port"),
		Host:     viper.GetString("rabbit.host"),
	}
}

func (cfg *Configurator) GetAMQPConnectionURL(rabbitCfg *RabbitMQConfig) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitCfg.Username, rabbitCfg.Password, rabbitCfg.Host, rabbitCfg.Port)
}
