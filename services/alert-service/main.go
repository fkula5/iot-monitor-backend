package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/skni-kod/iot-monitor-backend/internal/database"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent/alertrule"
)

// Struktura danych przychodząca z RabbitMQ (z data-processing)
type SensorData struct {
	SensorID  int64     `json:"sensor_id"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// Struktura zdarzenia wyjściowego (gdy wykryjemy alert)
type AlertEvent struct {
	AlertID   int       `json:"alert_id"`
	RuleID    int       `json:"rule_id"`
	SensorID  int64     `json:"sensor_id"`
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	// 1. Konfiguracja Bazy Danych
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("ALERT_SERVICE_DB_USER")
	dbPass := os.Getenv("ALERT_SERVICE_DB_PASSWORD")
	dbName := os.Getenv("ALERT_SERVICE_DB_NAME")

	// Fallbacki dla środowiska lokalnego (jeśli nie ma w .env)
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbUser == "" {
		dbUser = "alert_user"
	}
	if dbPass == "" {
		dbPass = "alertpassword"
	}
	if dbName == "" {
		dbName = "alert_db"
	}

	drv := database.NewDriver(dbHost, dbPort, dbUser, dbPass, dbName)
	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	// Automigracja schematu (tworzy tabele)
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	log.Println("Database connection established and schema migrated.")

	// 2. Konfiguracja RabbitMQ
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// 3. Deklaracja Giełd (Exchanges)
	// Giełda wejściowa (dane z czujników)
	err = ch.ExchangeDeclare("readings_exchange", "fanout", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare input exchange: %v", err)
	}

	// Giełda wyjściowa (wyzwolone alerty - dla dispatchera i api-gateway)
	err = ch.ExchangeDeclare("alerts_exchange", "fanout", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare output exchange: %v", err)
	}

	// 4. Kolejka dla tego serwisu
	q, err := ch.QueueDeclare(
		"alert_engine_queue", // Nazwa trwała (bo chcemy, żeby kolejka zbierała dane nawet jak serwis padnie)
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Bindowanie kolejki do giełdy z danymi
	err = ch.QueueBind(q.Name, "", "readings_exchange", false, nil)
	if err != nil {
		log.Fatalf("Failed to bind queue: %v", err)
	}

	// 5. Konsumpcja wiadomości
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer tag
		true,   // auto-ack (dla uproszczenia na start)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Println("Alert Service started. Waiting for sensor data...")

	// Pętla główna - "Silnik Reguł"
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			processMessage(client, ch, d.Body)
		}
	}()

	<-forever
}

// Funkcja Silnika: Sprawdza dane względem reguł
func processMessage(client *ent.Client, ch *amqp.Channel, body []byte) {
	var data SensorData
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		return
	}

	ctx := context.Background()

	// 1. Pobierz reguły dla tego konkretnego sensora
	// Optymalizacja: W prawdziwym systemie cache'owalibyśmy to w RAMie!
	rules, err := client.AlertRule.Query().
		Where(
			alertrule.SensorID(data.SensorID),
			alertrule.IsEnabled(true),
		).
		All(ctx)

	if err != nil {
		log.Printf("Error fetching rules: %v", err)
		return
	}

	// 2. Sprawdź każdą regułę
	for _, rule := range rules {
		isTriggered := false

		switch rule.ConditionType {
		case "GT": // Greater Than
			if data.Value > rule.Threshold {
				isTriggered = true
			}
		case "LT": // Less Than
			if data.Value < rule.Threshold {
				isTriggered = true
			}
		}

		if isTriggered {
			log.Printf("!!! ALERT TRIGGERED !!! Sensor: %d, Rule: %s, Value: %f", data.SensorID, rule.Name, data.Value)

			// 3. Zapisz Alert w bazie (Historia)
			savedAlert, err := client.Alert.Create().
				SetRule(rule).
				SetValue(data.Value).
				SetMessage(fmt.Sprintf("Rule '%s' violated: val %f", rule.Name, data.Value)).
				Save(ctx)

			if err != nil {
				log.Printf("Failed to save alert to DB: %v", err)
				continue
			}

			// 4. Wyślij zdarzenie do 'alerts_exchange' (dla Frontendu i Dispatchera)
			event := AlertEvent{
				AlertID:   savedAlert.ID,
				RuleID:    rule.ID,
				SensorID:  data.SensorID,
				Message:   savedAlert.Message,
				Value:     data.Value,
				Timestamp: time.Now(),
			}

			eventBody, _ := json.Marshal(event)

			err = ch.PublishWithContext(ctx,
				"alerts_exchange", // Fanout exchange
				"",
				false, false,
				amqp.Publishing{
					ContentType: "application/json",
					Body:        eventBody,
				},
			)

			if err != nil {
				log.Printf("Failed to publish alert event: %v", err)
			}
		}
	}
}
