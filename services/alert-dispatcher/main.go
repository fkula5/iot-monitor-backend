package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb_auth "github.com/skni-kod/iot-monitor-backend/internal/proto/auth"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
)

type AlertEvent struct {
        AlertID   int       `json:"alert_id"`
        RuleID    int       `json:"rule_id"`
        UserID    int64     `json:"user_id"`
        SensorID  int64     `json:"sensor_id"`
        Message   string    `json:"message"`
        Value     float64   `json:"value"`
        Timestamp time.Time `json:"timestamp"`
}
func main() {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	err := logger.Init(logger.Config{
		Level:       logLevel,
		Environment: environment,
		ServiceName: "alert-dispatcher",
		OutputPaths: []string{"stdout"},
	})
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("Starting Alert Dispatcher Service")

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Fatal("Failed to open a channel", zap.Error(err))
	}
	defer ch.Close()

	// Declare exchange and queue
	err = ch.ExchangeDeclare(
		"alerts_exchange",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Fatal("Failed to declare exchange", zap.Error(err))
	}

	q, err := ch.QueueDeclare(
		"alert_dispatcher_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Fatal("Failed to declare queue", zap.Error(err))
	}

	err = ch.QueueBind(
		q.Name,
		"",
		"alerts_exchange",
		false,
		nil,
	)
	if err != nil {
		logger.Fatal("Failed to declare queue bind", zap.Error(err))
	}

	authAddr := os.Getenv("AUTH_SERVICE_GRPC_ADDR")
	if authAddr == "" {
		authAddr = "localhost:50051"
	}

	authConn, err := grpc.NewClient(authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to Auth Service", zap.Error(err))
	}
	defer authConn.Close()

	authClient := pb_auth.NewAuthServiceClient(authConn)

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")

	if smtpFrom == "" {
		smtpFrom = "alerts@iot-monitor.local"
	}

	smtpPort := 587
	if smtpPortStr != "" {
		fmt.Sscanf(smtpPortStr, "%d", &smtpPort)
	}

	mailer := NewMailer(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom)

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Fatal("Failed to register a consumer", zap.Error(err))
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for d := range msgs {
			processAlert(d.Body, authClient, mailer)
		}
	}()

	logger.Info("Alert Dispatcher Service started. Waiting for alerts...")
	<-sigChan
	logger.Info("Shutting down Alert Dispatcher Service")
}

func processAlert(body []byte, authClient pb_auth.AuthServiceClient, mailer *Mailer) {
	var event AlertEvent
	if err := json.Unmarshal(body, &event); err != nil {
		logger.Error("Failed to unmarshal alert event", zap.Error(err))
		return
	}

	logger.Info("Received alert event",
		zap.Int("alert_id", event.AlertID),
		zap.Int64("user_id", event.UserID),
		zap.String("message", event.Message),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userRes, err := authClient.GetUser(ctx, &pb_auth.GetUserRequest{Id: event.UserID})
	if err != nil {
	        logger.Error("Failed to fetch user details", zap.Int64("user_id", event.UserID), zap.Error(err))
	        return
	}
	logger.Info("Dispatching alert to user",
		zap.String("email", userRes.User.Email),
		zap.String("username", userRes.User.Username),
	)

	if err := mailer.SendAlertEmail(userRes.User.Email, event); err != nil {
		logger.Error("Failed to send alert email",
			zap.String("to", userRes.User.Email),
			zap.Error(err),
		)
		return
	}

	logger.Info("Successfully sent alert email", zap.String("to", userRes.User.Email))
}
