package types

import (
	"time"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
)

type CreateSensorRequest struct {
	Name         string `json:"name"`
	Location     string `json:"location"`
	Description  string `json:"description"`
	SensorTypeId int64  `json:"sensor_type_id"`
}

type UpdateSensorRequest struct {
	Name         *string `json:"name,omitempty"`
	Location     *string `json:"location,omitempty"`
	Description  *string `json:"description,omitempty"`
	SensorTypeId *int64  `json:"sensor_type_id,omitempty"`
	Active       *bool   `json:"active,omitempty"`
}

type SensorResponse struct {
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	Location    string              `json:"location"`
	Description string              `json:"description"`
	Active      bool                `json:"active"`
	LastUpdated *time.Time          `json:"last_updated,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	SensorType  *SensorTypeResponse `json:"sensor_type,omitempty"`
}

type SensorTypeResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Model        string    `json:"model"`
	Manufacturer string    `json:"manufacturer,omitempty"`
	Description  string    `json:"description,omitempty"`
	Unit         string    `json:"unit,omitempty"`
	MinValue     float32   `json:"min_value,omitempty"`
	MaxValue     float32   `json:"max_value,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

func MapSensorFromProto(s *pb.Sensor) SensorResponse {
	response := SensorResponse{
		ID:          s.Id,
		Name:        s.Name,
		Location:    s.Location,
		Description: s.Description,
		Active:      s.Active,
		CreatedAt:   s.CreatedAt.AsTime(),
		UpdatedAt:   s.UpdatedAt.AsTime(),
	}

	if s.LastUpdated != nil {
		t := s.LastUpdated.AsTime()
		response.LastUpdated = &t
	}

	if s.SensorType != nil {
		response.SensorType = &SensorTypeResponse{
			ID:           s.SensorType.Id,
			Name:         s.SensorType.Name,
			Model:        s.SensorType.Model,
			Manufacturer: s.SensorType.Manufacturer,
			Description:  s.SensorType.Description,
			Unit:         s.SensorType.Unit,
			MinValue:     s.SensorType.MinValue,
			MaxValue:     s.SensorType.MaxValue,
			CreatedAt:    s.SensorType.CreatedAt.AsTime(),
		}
	} else if s.SensorTypeId > 0 {
		response.SensorType = &SensorTypeResponse{
			ID: s.SensorTypeId,
		}
	}

	return response
}
