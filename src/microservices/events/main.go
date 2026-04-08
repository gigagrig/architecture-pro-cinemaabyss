package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

const (
	movieTopic   = "movie-events"
	userTopic    = "user-events"
	paymentTopic = "payment-events"
)

type service struct {
	producer sarama.SyncProducer
	consumer sarama.Consumer
}

type event struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp string      `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

type eventResponse struct {
	Status    string `json:"status"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
	Event     event  `json:"event"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type movieEvent struct {
	MovieID     int      `json:"movie_id"`
	Title       string   `json:"title"`
	Action      string   `json:"action"`
	UserID      int      `json:"user_id,omitempty"`
	Rating      float64  `json:"rating,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	Description string   `json:"description,omitempty"`
}

type userEvent struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username,omitempty"`
	Email     string `json:"email,omitempty"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
}

type paymentEvent struct {
	PaymentID  int     `json:"payment_id"`
	UserID     int     `json:"user_id"`
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"`
	Timestamp  string  `json:"timestamp"`
	MethodType string  `json:"method_type,omitempty"`
}

func main() {
	port := envOrDefault("PORT", "8082")
	brokers := splitBrokers(envOrDefault("KAFKA_BROKERS", "localhost:9092"))

	svc, err := newService(brokers)
	if err != nil {
		log.Fatalf("failed to initialize kafka clients: %v", err)
	}
	defer func() {
		if err := svc.producer.Close(); err != nil {
			log.Printf("failed to close producer: %v", err)
		}
		if err := svc.consumer.Close(); err != nil {
			log.Printf("failed to close consumer: %v", err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	svc.startConsumers(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/events/health", healthHandler)
	mux.HandleFunc("/api/events/movie", svc.handleMovieEvent)
	mux.HandleFunc("/api/events/user", svc.handleUserEvent)
	mux.HandleFunc("/api/events/payment", svc.handlePaymentEvent)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("failed to shutdown events service: %v", err)
		}
	}()

	log.Printf("starting events service on port %s", port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func newService(brokers []string) (*service, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_7_0_0
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Consumer.Return.Errors = true

	var (
		producer sarama.SyncProducer
		consumer sarama.Consumer
		err      error
	)

	for attempt := 1; attempt <= 30; attempt++ {
		producer, err = sarama.NewSyncProducer(brokers, config)
		if err == nil {
			consumer, err = sarama.NewConsumer(brokers, config)
		}
		if err == nil {
			return &service{producer: producer, consumer: consumer}, nil
		}

		if producer != nil {
			_ = producer.Close()
		}
		log.Printf("waiting for kafka (attempt %d/30): %v", attempt, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("kafka is unavailable after retries: %w", err)
}

func (s *service) startConsumers(ctx context.Context) {
	for _, topic := range []string{movieTopic, userTopic, paymentTopic} {
		partitionConsumer, err := s.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
		if err != nil {
			log.Printf("failed to start consumer for topic %s: %v", topic, err)
			continue
		}

		go func(topic string, partitionConsumer sarama.PartitionConsumer) {
			defer func() {
				if err := partitionConsumer.Close(); err != nil {
					log.Printf("failed to close consumer for topic %s: %v", topic, err)
				}
			}()

			for {
				select {
				case msg := <-partitionConsumer.Messages():
					if msg == nil {
						continue
					}
					log.Printf("consumed topic=%s partition=%d offset=%d payload=%s", topic, msg.Partition, msg.Offset, string(msg.Value))
				case err := <-partitionConsumer.Errors():
					if err != nil {
						log.Printf("consumer error for topic %s: %v", topic, err)
					}
				case <-ctx.Done():
					return
				}
			}
		}(topic, partitionConsumer)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"status": true})
}

func (s *service) handleMovieEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var payload movieEvent
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if payload.MovieID <= 0 || strings.TrimSpace(payload.Title) == "" || strings.TrimSpace(payload.Action) == "" {
		writeError(w, http.StatusBadRequest, "movie_id, title and action are required")
		return
	}

	evt := event{
		ID:        fmt.Sprintf("movie-%d-%d", payload.MovieID, time.Now().UnixNano()),
		Type:      "movie",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Payload:   payload,
	}
	s.publishEvent(w, evt, movieTopic)
}

func (s *service) handleUserEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var payload userEvent
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if payload.UserID <= 0 || strings.TrimSpace(payload.Action) == "" || strings.TrimSpace(payload.Timestamp) == "" {
		writeError(w, http.StatusBadRequest, "user_id, action and timestamp are required")
		return
	}
	if _, err := time.Parse(time.RFC3339, payload.Timestamp); err != nil {
		writeError(w, http.StatusBadRequest, "timestamp must be RFC3339")
		return
	}

	evt := event{
		ID:        fmt.Sprintf("user-%d-%d", payload.UserID, time.Now().UnixNano()),
		Type:      "user",
		Timestamp: payload.Timestamp,
		Payload:   payload,
	}
	s.publishEvent(w, evt, userTopic)
}

func (s *service) handlePaymentEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var payload paymentEvent
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if payload.PaymentID <= 0 || payload.UserID <= 0 || payload.Amount <= 0 || strings.TrimSpace(payload.Status) == "" || strings.TrimSpace(payload.Timestamp) == "" {
		writeError(w, http.StatusBadRequest, "payment_id, user_id, amount, status and timestamp are required")
		return
	}
	if _, err := time.Parse(time.RFC3339, payload.Timestamp); err != nil {
		writeError(w, http.StatusBadRequest, "timestamp must be RFC3339")
		return
	}

	evt := event{
		ID:        fmt.Sprintf("payment-%d-%d", payload.PaymentID, time.Now().UnixNano()),
		Type:      "payment",
		Timestamp: payload.Timestamp,
		Payload:   payload,
	}
	s.publishEvent(w, evt, paymentTopic)
}

func (s *service) publishEvent(w http.ResponseWriter, evt event, topic string) {
	body, err := json.Marshal(evt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to serialize event")
		return
	}

	partition, offset, err := s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(evt.ID),
		Value: sarama.ByteEncoder(body),
	})
	if err != nil {
		log.Printf("failed to publish event to topic %s: %v", topic, err)
		writeError(w, http.StatusInternalServerError, "failed to publish event")
		return
	}

	writeJSON(w, http.StatusCreated, eventResponse{
		Status:    "success",
		Partition: partition,
		Offset:    offset,
		Event:     evt,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func splitBrokers(raw string) []string {
	parts := strings.Split(raw, ",")
	brokers := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			brokers = append(brokers, part)
		}
	}
	if len(brokers) == 0 {
		return []string{"localhost:9092"}
	}
	return brokers
}
