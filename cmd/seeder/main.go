package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/skni-kod/iot-monitor-backend/internal/auth"
	"github.com/skni-kod/iot-monitor-backend/internal/database"
	authEnt "github.com/skni-kod/iot-monitor-backend/services/auth/ent"
	sensorEnt "github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
)

type UserSeed struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type SensorTypeSeed struct {
	Name         string  `json:"name"`
	Model        string  `json:"model"`
	Manufacturer string  `json:"manufacturer"`
	Description  string  `json:"description"`
	Unit         string  `json:"unit"`
	MinValue     float64 `json:"min_value"`
	MaxValue     float64 `json:"max_value"`
}

type SensorSeed struct {
	Name         string `json:"name"`
	Location     string `json:"location"`
	Description  string `json:"description"`
	Active       bool   `json:"active"`
	UserID       int64  `json:"user_id"`
	SensorTypeID int    `json:"sensor_type_id"`
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	userDbName := os.Getenv("AUTH_SERVICE_DB_NAME")
	sensorDbName := os.Getenv("SENSOR_SERVICE_DB_NAME")

	userDriver := database.NewDriver(dbHost, dbPort, dbUser, dbPass, userDbName)
	sensorDriver := database.NewDriver(dbHost, dbPort, dbUser, dbPass, sensorDbName)

	userClient := authEnt.NewClient(authEnt.Driver(userDriver))

	sensorClient := sensorEnt.NewClient(sensorEnt.Driver(sensorDriver))

	defer userClient.Close()
	defer sensorClient.Close()

	userJson, err := os.Open("data/users.json")
	if err != nil {
		panic(err)
	}
	defer userJson.Close()

	sensorTypeJson, err := os.Open("data/sensor_types.json")
	if err != nil {
		panic(err)
	}
	defer sensorTypeJson.Close()

	sensorJson, err := os.Open("data/sensors.json")
	if err != nil {
		panic(err)
	}
	defer sensorJson.Close()

	var users []UserSeed
	var sensorTypes []SensorTypeSeed
	var sensors []SensorSeed

	if err := json.NewDecoder(userJson).Decode(&users); err != nil {
		panic(err)
	}

	if err := json.NewDecoder(sensorTypeJson).Decode(&sensorTypes); err != nil {
		panic(err)
	}

	if err := json.NewDecoder(sensorJson).Decode(&sensors); err != nil {
		panic(err)
	}

	for _, u := range users {
		hashedPassword, err := auth.NewPasswordService().HashPassword(u.Password)
		if err != nil {
			panic(err)
		}
		_, err = userClient.User.
			Create().
			SetUsername(u.Username).
			SetEmail(u.Email).
			SetPasswordHash(hashedPassword).
			SetFirstName(u.FirstName).
			SetLastName(u.LastName).
			Save(context.Background())
		if err != nil {
			panic(err)
		}
	}

	for _, st := range sensorTypes {
		_, err := sensorClient.SensorType.
			Create().
			SetName(st.Name).
			SetModel(st.Model).
			SetManufacturer(st.Manufacturer).
			SetDescription(st.Description).
			SetUnit(st.Unit).
			SetMinValue(st.MinValue).
			SetMaxValue(st.MaxValue).
			Save(context.Background())
		if err != nil {
			panic(err)
		}
	}

	for _, s := range sensors {
		_, err := sensorClient.Sensor.
			Create().
			SetName(s.Name).
			SetLocation(s.Location).
			SetDescription(s.Description).
			SetActive(s.Active).
			SetUserID(s.UserID).
			SetTypeID(s.SensorTypeID).
			Save(context.Background())
		if err != nil {
			panic(err)
		}
	}
}
