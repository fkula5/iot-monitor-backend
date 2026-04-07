package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/skni-kod/iot-monitor-backend/internal/database"
	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/alert_service"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent/alertrule"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/handlers"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/service"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/storage"
)

type SensorData struct {
	SensorID  int64     `json:"sensor_id"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type AlertEvent struct {
	AlertID   int       `json:"alert_id"`
	RuleID    int       `json:"rule_id"`
	UserID    int64     `json:"user_id"`
	SensorID  int64     `json:"sensor_id"`
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

func getEnvOrFail(key string) string {
	value := os.Getenv(key)
	if value == "" {
		logger.Fatal("Environment variable not set", zap.String("key", key))
	}
	return value
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
		ServiceName: "alert-service",
		OutputPaths: []string{"stdout"},
	})
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Alert Service",
		zap.String("environment", environment),
		zap.String("log_level", logLevel),
	)

	dbHost := getEnvOrFail("DB_HOST")
	dbPort := getEnvOrFail("DB_PORT")
	dbUser := getEnvOrFail("ALERT_SERVICE_DB_USER")
	dbPass := getEnvOrFail("ALERT_SERVICE_DB_PASSWORD")
	dbName := getEnvOrFail("ALERT_SERVICE_DB_NAME")
	grpcPort := getEnvOrFail("ALERT_SERVICE_GRPC_PORT")

	drv := database.NewDriver(dbHost, dbPort, dbUser, dbPass, dbName)
	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	if err := client.Schema.Create(context.Background()); err != nil {
		logger.Fatal("failed creating schema resources", zap.Error(err))
	}
	logger.Info("Database connection established and schema migrated")

	alertStorage := storage.NewAlertStorage(client)
	alertRuleStorage := storage.NewAlertRuleStorage(client)
	alertService := service.NewAlertService(alertStorage)
	alertRuleService := service.NewAlertRuleService(alertRuleStorage)
	handler := handlers.NewAlertGrpcHandler(alertService, alertRuleService)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Fatal("failed to listen", zap.String("port", grpcPort), zap.Error(err))
	}

	s := grpc.NewServer()
	pb.RegisterAlertServiceServer(s, handler)

	go func() {
		logger.Info("gRPC server listening", zap.String("port", grpcPort))
		if err := s.Serve(lis); err != nil {
			logger.Fatal("failed to serve gRPC", zap.Error(err))
		}
	}()

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
		logger.Fatal("Failed to open a RabbitMQ channel", zap.Error(err))
	}
	defer ch.Close()

	setupRabbitMQ(ch)

	msgs, err := ch.Consume("alert_engine_queue", "", true, false, false, false, nil)
	if err != nil {
		logger.Fatal("Failed to register a consumer", zap.Error(err))
	}

	logger.Info("Alert Service started. Waiting for sensor data...")

	for d := range msgs {
		processMessage(client, ch, d.Body)
	}
}

func setupRabbitMQ(ch *amqp.Channel) {
	err := ch.ExchangeDeclare("readings_exchange", "fanout", true, false, false, false, nil)
	if err != nil {
		logger.Fatal("Failed to declare readings_exchange", zap.Error(err))
	}

	err = ch.ExchangeDeclare("alerts_exchange", "fanout", true, false, false, false, nil)
	if err != nil {
		logger.Fatal("Failed to declare alerts_exchange", zap.Error(err))
	}

	q, err := ch.QueueDeclare("alert_engine_queue", true, false, false, false, nil)
	if err != nil {
		logger.Fatal("Failed to declare alert_engine_queue", zap.Error(err))
	}

	err = ch.QueueBind(q.Name, "", "readings_exchange", false, nil)
	if err != nil {
		logger.Fatal("Failed to bind queue", zap.Error(err))
	}
}

func processMessage(client *ent.Client, ch *amqp.Channel, body []byte) {
	var data SensorData
	if err := json.Unmarshal(body, &data); err != nil {
		logger.Error("Error decoding JSON", zap.Error(err))
		return
	}

	ctx := context.Background()
	rules, err := client.AlertRule.Query().
		Where(alertrule.SensorID(data.SensorID), alertrule.IsEnabled(true)).
		All(ctx)

	if err != nil {
		logger.Error("Error fetching rules", zap.Int64("sensor_id", data.SensorID), zap.Error(err))
		return
	}

	for _, rule := range rules {
		if isTriggered(rule, data.Value) {
			logger.Info("Alert triggered",
				zap.Int64("sensor_id", data.SensorID),
				zap.String("rule_name", rule.Name),
				zap.Float64("value", data.Value),
				zap.Float64("threshold", rule.Threshold),
			)

			savedAlert, err := client.Alert.Create().
				SetRule(rule).
				SetUserID(rule.UserID).
				SetValue(data.Value).
				SetMessage(fmt.Sprintf("Rule '%s' violated: val %f", rule.Name, data.Value)).
				Save(ctx)

			if err != nil {
				logger.Error("Failed to save alert to DB", zap.Error(err))
				continue
			}

			publishAlert(ch, ctx, savedAlert, rule, data.Value)
		}
	}
}

func isTriggered(rule *ent.AlertRule, value float64) bool {
	if rule.ConditionType == "GT" {
		return value > rule.Threshold
	}
	if rule.ConditionType == "LT" {
		return value < rule.Threshold
	}
	return false
}

func publishAlert(ch *amqp.Channel, ctx context.Context, a *ent.Alert, rule *ent.AlertRule, val float64) {
	event := AlertEvent{
		AlertID:   a.ID,
		RuleID:    rule.ID,
		UserID:    rule.UserID,
		SensorID:  rule.SensorID,
		Message:   a.Message,
		Value:     val,
		Timestamp: time.Now(),
	}
	body, _ := json.Marshal(event)
	err := ch.PublishWithContext(ctx, "alerts_exchange", "", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		logger.Error("Failed to publish alert event", zap.Error(err))
	}
}
