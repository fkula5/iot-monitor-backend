package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "fk"
	password = "2w3e4r5tK$"
	dbname   = "sensors"
)

func Open(databaseUrl string) *ent.Client {
	db, err := sql.Open("postgres", databaseUrl)
	if err != nil {
		log.Fatal(err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	return ent.NewClient(ent.Driver(drv))
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	client := Open(psqlInfo)
	defer client.Close()

	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	fmt.Println("Successfully initialized Ent client and created schema!")
}
