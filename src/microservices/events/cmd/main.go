package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/swa-project-sprint-2/src/microservices/events/internal/adapter/produser"
	"github.com/swa-project-sprint-2/src/microservices/events/internal/controller/consumer"
	"github.com/swa-project-sprint-2/src/microservices/events/internal/lib/config"
	"golang.org/x/sync/errgroup"
)

// Models
type EventMovie struct {
	MovieID     int      `json:"movie_id"`
	Title       string   `json:"title"`
	Action      string   `json:"action"`
	UserID      int      `json:"user_id"`
	Rating      float64  `json:"rating"`
	Description string   `json:"description"`
	Genres      []string `json:"genres"`
}

type EventUser struct {
	UserID    int       `json:"user_id"`
	UserName  string    `json:"username"`
	Email     string    `json:"email"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

type EventPayment struct {
	PaymentID int       `json:"payment_id"`
	UserID    int       `json:"user_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"username"`
	MetodType string    `json:"metod_type"`
	Timestamp time.Time `json:"timestamp"`
}

type kafkaPayload struct {
	RequestID    string        `json:"request_id"`
	EventMovie   *EventMovie   `json:"event_movie"`
	EventUser    *EventUser    `json:"event_user"`
	EventPayment *EventPayment `json:"event_payment"`
}

var (
	cfg           config.Config
	kafkaProducer *produser.KafkaProducer
)

func init() {

	flag.StringVar(&cfg.Port, "port", "8082", "port")
	flag.StringVar(&cfg.KafkaBrokers, "kafka-brokers", "localhost:9092", "kafka brokers")

	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		cfg.Port = envPort
	}

	if envKafkaBrokers := os.Getenv("KAFKA_BROKERS"); envKafkaBrokers != "" {
		cfg.KafkaBrokers = envKafkaBrokers
	}

}

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	kafkaProducer, err := produser.NewKafkaProducer(cfg.KafkaBrokers)
	if err != nil {
		logrus.Fatalf("Error initializing kafka: %v", err)
	}
	defer kafkaProducer.Close()

	logrus.Info("Successful connect from kafka")

	// Start server
	port := cfg.Port
	logrus.Infof("Starting events microservice on port %s", port)

	server := &http.Server{
		Addr: fmt.Sprintf("localhost:%s", cfg.Port),
	}

	// Создание анонимной функции для передачи дополнительного параметра
	handleEventMovieWithKafka := func(w http.ResponseWriter, r *http.Request) {
		handleEventMovie(w, r, kafkaProducer)
	}
	handleEventUserWithKafka := func(w http.ResponseWriter, r *http.Request) {
		handleEventUser(w, r, kafkaProducer)
	}
	handleEventPaymentWithKafka := func(w http.ResponseWriter, r *http.Request) {
		handleEventPayment(w, r, kafkaProducer)
	}

	// Set up HTTP routes
	http.HandleFunc("/api/events/movie", handleEventMovieWithKafka)
	http.HandleFunc("/api/events/user", handleEventUserWithKafka)
	http.HandleFunc("/api/events/payment", handleEventPaymentWithKafka)
	http.HandleFunc("/api/events/health", handleHealth)

	errg, ctx := errgroup.WithContext(ctx)
	errg.Go(func() error {
		return server.ListenAndServe()
	})

	consumerMovieEvents, err := consumer.NewKafkaConsumer(cfg.KafkaBrokers, "movie-events")
	if err != nil {
		logrus.Fatalf("failed to connect to kafka: %#v", err)
	}

	consumerUserEvents, err := consumer.NewKafkaConsumer(cfg.KafkaBrokers, "user-events")
	if err != nil {
		logrus.Fatalf("failed to connect to kafka: %#v", err)
	}

	consumerPaymentEvents, err := consumer.NewKafkaConsumer(cfg.KafkaBrokers, "payment-events")
	if err != nil {
		logrus.Fatalf("failed to connect to kafka: %#v", err)
	}

	errg.Go(func() error {
		return consumerMovieEvents.ListenForMessages(ctx, 10)
	})

	errg.Go(func() error {
		return consumerUserEvents.ListenForMessages(ctx, 10)
	})

	errg.Go(func() error {
		return consumerPaymentEvents.ListenForMessages(ctx, 10)
	})

	<-ctx.Done()
	err = server.Shutdown(ctx)
	if err != nil {
		logrus.Errorf("server shutdown with error: %v", err)
	}
	logrus.Infof("Server is graceful shutdown...")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

// Movie handlers
func handleEventMovie(w http.ResponseWriter, r *http.Request, kafkaProducer *produser.KafkaProducer) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	createEventMovie(w, r, kafkaProducer)
}

// Movie handlers
func handleEventUser(w http.ResponseWriter, r *http.Request, kafkaProducer *produser.KafkaProducer) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	createEventUser(w, r, kafkaProducer)
}

func handleEventPayment(w http.ResponseWriter, r *http.Request, kafkaProducer *produser.KafkaProducer) {

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	createEventPayment(w, r, kafkaProducer)
}

func createEventMovie(w http.ResponseWriter, r *http.Request, kafkaProducer *produser.KafkaProducer) {

	var m EventMovie

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := uuid.New().String()

	payload := kafkaPayload{
		RequestID:  id,
		EventMovie: &m,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = kafkaProducer.SendMessage("movie-events", []byte(id), data, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func createEventUser(w http.ResponseWriter, r *http.Request, kafkaProducer *produser.KafkaProducer) {

	var m EventUser
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := uuid.New().String()

	payload := kafkaPayload{
		RequestID: id,
		EventUser: &m,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = kafkaProducer.SendMessage("user-events", []byte(id), data, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func createEventPayment(w http.ResponseWriter, r *http.Request, kafkaProducer *produser.KafkaProducer) {

	var m EventPayment
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := uuid.New().String()

	payload := kafkaPayload{
		RequestID:    id,
		EventPayment: &m,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = kafkaProducer.SendMessage("payment-events", []byte(id), data, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
