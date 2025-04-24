package database

import (
	"database/sql"
	"fmt"
	"log"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
)

type DatabaseClient struct {
	Client *ent.Client
}

func New(host, port, user, password, dbname string) *DatabaseClient {
	connStr := fmt.Sprintf("host=%s port=%s user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to establish database connectio: %v", err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)

	dbClient := &DatabaseClient{Client: ent.NewClient(ent.Driver(drv))}

	return dbClient
}
