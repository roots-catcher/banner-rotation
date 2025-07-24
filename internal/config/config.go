package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Kafka KafkaConfig
}

type KafkaConfig struct {
	Brokers     string
	TopicEvents string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/rotation/")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Переопределение переменными окружения
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		cfg.Kafka.Brokers = brokers
	}
	if topic := os.Getenv("KAFKA_TOPIC"); topic != "" {
		cfg.Kafka.TopicEvents = topic
	}

	return &cfg, nil
}
