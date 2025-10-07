package database

import (
	"database/sql"
	"fmt"
	"log"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
	userEnt "github.com/skni-kod/iot-monitor-backend/services/auth/ent"
	sensorEnt "github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
)

func NewSensorDB(host, port, user, password, dbname string) *sensorEnt.Client {
	connStr := fmt.Sprintf("host=%s port=%s user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to establish database connectio: %v", err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)

	return sensorEnt.NewClient(sensorEnt.Driver(drv))
}

func NewAuthDB(host, port, user, password, dbname string) *userEnt.Client {
	connStr := fmt.Sprintf("host=%s port=%s user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to establish database connectio: %v", err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	return userEnt.NewClient(userEnt.Driver(drv))
}
