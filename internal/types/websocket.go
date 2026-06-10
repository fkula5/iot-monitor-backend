package types

import "time"

type ReadingMessage struct {
	SensorID   int64     `json:"sensor_id"`
	Value      float32   `json:"value"`
	Timestamp  time.Time `json:"timestamp"`
	SensorName string    `json:"sensor_name"`
	Location   string    `json:"location"`
	Unit       string    `json:"unit"`
}

type SubscribeMessage struct {
	Type      string  `json:"type"`
	SensorIDs []int64 `json:"sensor_ids"`
}

type StoreReadingRequest struct {
	SensorID  int64     `json:"sensor_id"`
	Value     float32   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type ReadingResponse struct {
	SensorID  int64     `json:"sensor_id"`
	Value     float32   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type AlertMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
