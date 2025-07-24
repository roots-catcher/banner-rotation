package main

import (
	"banner-rotation/internal/api"
	"banner-rotation/internal/app"
	"banner-rotation/internal/config"
	"banner-rotation/internal/kafka"
	"banner-rotation/internal/storage/postgres"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	segmentio_kafka "github.com/segmentio/kafka-go"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Подключение к PostgreSQL
	store, err := postgres.New(os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("Error closing store: %v", err)
		}
	}()

	// Создание продюсера Kafka
	var producer kafka.ProducerInterface
	if cfg.Kafka.Brokers != "" {
		kafkaWriter := &segmentio_kafka.Writer{
			Addr:         segmentio_kafka.TCP(strings.Split(cfg.Kafka.Brokers, ",")...),
			Topic:        cfg.Kafka.TopicEvents,
			Balancer:     &segmentio_kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
		}
		producer = kafka.NewProducer(kafkaWriter)
		defer func() {
			if err := producer.Close(); err != nil {
				log.Printf("Error closing producer: %v", err)
			}
		}()
	} else {
		log.Println("Kafka brokers not configured, producer disabled")
		producer = nil
	}

	// Инициализация сервиса
	bandit := app.NewBandit(store, producer)

	// Создание и запуск API сервера
	apiServer := api.NewServer(bandit)
	go func() {
		log.Println("Starting API server on :8080")
		if err := apiServer.Start(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("API server failed: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	<-ctx.Done()
	log.Println("Shutting down..")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("API server shutdown error: %v", err)
	}
	log.Println("Server exited")
}
