package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/joho/godotenv"

	"github.com/skni-kod/iot-monitor-backend/internal/auth"
	"github.com/skni-kod/iot-monitor-backend/internal/database"
	alertEnt "github.com/skni-kod/iot-monitor-backend/services/alert-service/ent"
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

type AlertRuleSeed struct {
	Name          string  `json:"name"`
	UserID        int64   `json:"user_id"`
	SensorID      int64   `json:"sensor_id"`
	ConditionType string  `json:"condition_type"`
	Threshold     float64 `json:"threshold"`
	Description   string  `json:"description"`
	IsEnabled     bool    `json:"is_enabled"`
}

type AlertSeed struct {
	UserID    int64   `json:"user_id"`
	RuleID    int     `json:"rule_id"`
	Value     float64 `json:"value"`
	Message   string  `json:"message"`
	IsRead    bool    `json:"is_read"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	userDbName := os.Getenv("AUTH_SERVICE_DB_NAME")
	sensorDbName := os.Getenv("SENSOR_SERVICE_DB_NAME")
	alertDbName := os.Getenv("ALERT_SERVICE_DB_NAME")

	userDriver := database.NewDriver(dbHost, dbPort, dbUser, dbPass, userDbName)
	sensorDriver := database.NewDriver(dbHost, dbPort, dbUser, dbPass, sensorDbName)
	alertDriver := database.NewDriver(dbHost, dbPort, dbUser, dbPass, alertDbName)

	userClient := authEnt.NewClient(authEnt.Driver(userDriver))
	sensorClient := sensorEnt.NewClient(sensorEnt.Driver(sensorDriver))
	alertClient := alertEnt.NewClient(alertEnt.Driver(alertDriver))

	defer userClient.Close()
	defer sensorClient.Close()
	defer alertClient.Close()

	userJson, err := os.Open("cmd/seeder/data/users.json")
	if err != nil {
		panic(err)
	}
	defer userJson.Close()

	sensorTypeJson, err := os.Open("cmd/seeder/data/sensortypes.json")
	if err != nil {
		panic(err)
	}
	defer sensorTypeJson.Close()

	sensorJson, err := os.Open("cmd/seeder/data/sensors.json")
	if err != nil {
		panic(err)
	}
	defer sensorJson.Close()

	alertRuleJson, err := os.Open("cmd/seeder/data/alertrules.json")
	if err != nil {
		panic(err)
	}
	defer alertRuleJson.Close()

	alertJson, err := os.Open("cmd/seeder/data/alerts.json")
	if err != nil {
		panic(err)
	}
	defer alertJson.Close()

	var users []UserSeed
	var sensorTypes []SensorTypeSeed
	var sensors []SensorSeed
	var alertRules []AlertRuleSeed
	var alerts []AlertSeed

	if err := json.NewDecoder(userJson).Decode(&users); err != nil {
		panic(err)
	}

	if err := json.NewDecoder(sensorTypeJson).Decode(&sensorTypes); err != nil {
		panic(err)
	}

	if err := json.NewDecoder(sensorJson).Decode(&sensors); err != nil {
		panic(err)
	}

	if err := json.NewDecoder(alertRuleJson).Decode(&alertRules); err != nil {
		panic(err)
	}

	if err := json.NewDecoder(alertJson).Decode(&alerts); err != nil {
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

	for _, ar := range alertRules {
		_, err := alertClient.AlertRule.
			Create().
			SetName(ar.Name).
			SetUserID(ar.UserID).
			SetSensorID(ar.SensorID).
			SetConditionType(ar.ConditionType).
			SetThreshold(ar.Threshold).
			SetDescription(ar.Description).
			SetIsEnabled(ar.IsEnabled).
			Save(context.Background())
		if err != nil {
			panic(err)
		}
	}

	for _, a := range alerts {
		_, err := alertClient.Alert.
			Create().
			SetUserID(a.UserID).
			SetRuleID(a.RuleID).
			SetValue(a.Value).
			SetMessage(a.Message).
			SetIsRead(a.IsRead).
			Save(context.Background())
		if err != nil {
			panic(err)
		}
	}
}
