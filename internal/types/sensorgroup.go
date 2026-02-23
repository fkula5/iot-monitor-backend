package types

import (
	"time"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
)

type CreateGroupRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	Icon        string  `json:"icon"`
	SensorIDs   []int64 `json:"sensor_ids"`
}

type UpdateGroupRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	Icon        string  `json:"icon"`
	SensorIDs   []int64 `json:"sensor_ids"`
}

type AddSensorsRequest struct {
	SensorIDs []int64 `json:"sensor_ids"`
}

type SensorGroupResponse struct {
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Color       string           `json:"color"`
	Icon        string           `json:"icon"`
	UserID      int64            `json:"user_id,omitempty"`
	SensorCount int              `json:"sensor_count"`
	Sensors     []SensorResponse `json:"sensors,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

func MapSensorGroupFromProto(sg *pb.SensorGroup, mappedSensors []SensorResponse) SensorGroupResponse {
	return SensorGroupResponse{
		ID:          sg.Id,
		Name:        sg.Name,
		Description: sg.Description,
		Color:       sg.Color,
		Icon:        sg.Icon,
		UserID:      sg.UserId,
		SensorCount: len(sg.SensorIds),
		Sensors:     mappedSensors,
		CreatedAt:   sg.CreatedAt.AsTime(),
		UpdatedAt:   sg.UpdatedAt.AsTime(),
	}
}
