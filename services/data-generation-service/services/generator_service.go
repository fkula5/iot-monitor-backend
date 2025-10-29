package services

import (
	"context"
	"log"
	"math/rand/v2"
	"time"

	"github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
)

type IGeneratorService interface {
	Start(ctx context.Context) error
	Stop() error
}

type GeneratorService struct {
	sensorClient        sensor_service.SensorServiceClient
	generationInterval  time.Duration
	lastValues          map[int64]float64
	stopChan            chan struct{}
	generationInProcess bool
}

func NewGeneratorService(sensorClient sensor_service.SensorServiceClient) IGeneratorService {
	return &GeneratorService{
		sensorClient:        sensorClient,
		generationInterval:  5 * time.Second,
		lastValues:          make(map[int64]float64),
		stopChan:            make(chan struct{}),
		generationInProcess: false,
	}
}

func (g *GeneratorService) Start(ctx context.Context) error {
	if g.generationInProcess {
		return nil
	}

	g.generationInProcess = true

	ticker := time.NewTicker(g.generationInterval)

	go func() {
		defer ticker.Stop()

		log.Println("Started data generator")

		g.generateData(ctx)

		for {
			select {
			case <-ticker.C:
				g.generateData(ctx)

			case <-g.stopChan:
				log.Println("Stopping data generator")
				return
			case <-ctx.Done():
				log.Println("Context cancelled, stopping data generator")
				return
			}
		}
	}()

	return nil
}

func (g *GeneratorService) Stop() error {
	if !g.generationInProcess {
		return nil
	}

	g.stopChan <- struct{}{}
	g.generationInProcess = false
	return nil
}

func (g *GeneratorService) generateData(ctx context.Context) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sensors, err := g.sensorClient.ListSensors(ctxTimeout, &sensor_service.ListSensorsRequest{})
	if err != nil {
		log.Printf("Error fetching sensors: %v", err)
		return
	}

	activeSensors := 0
	for _, sensor := range sensors.Sensors {
		if !sensor.Active {
			continue
		}
		activeSensors++

		sensorDetails, err := g.sensorClient.GetSensor(ctxTimeout, &sensor_service.GetSensorRequest{Id: int64(sensor.Id)})
		if err != nil {
			log.Printf("Error fetching sensor details for sensor %d: %v", sensor.Id, err)
			continue
		}

		if sensorDetails == nil {
			log.Printf("Sensor %d not found", sensor.Id)
			continue
		}

		sensorTypeId := sensorDetails.Sensor.SensorTypeId
		sensorType, err := g.sensorClient.GetSensorType(ctxTimeout, &sensor_service.GetSensorTypeRequest{Id: int64(sensorTypeId)})
		if err != nil {
			log.Printf("Error fetching sensor type for sensor %d: %v", sensor.Id, err)
			continue
		}

		if sensorType == nil {
			log.Printf("Sensor type %d not found for sensor %d", sensorTypeId, sensor.Id)
			continue
		}

		minValue := sensorType.SensorType.MinValue
		maxValue := sensorType.SensorType.MaxValue
		unit := sensorType.SensorType.Unit

		if minValue == maxValue {
			minValue = 0.0
			maxValue = 100.0
		}

		generatedValue := g.generateRealisticValue(sensor.Id, minValue, maxValue)

		g.lastValues[sensor.Id] = generatedValue

		log.Printf("Generated data: Sensor ID=%d, Name=%s, Location=%s, Value=%.2f %s",
			sensor.Id, sensor.Name, sensor.Location, generatedValue, unit)
	}

	log.Printf("Generated data for %d active sensors", activeSensors)
}

func (g *GeneratorService) generateRealisticValue(sensorID int64, minValue, maxValue float32) float64 {
	min := float64(minValue)
	max := float64(maxValue)
	range_ := max - min

	if lastValue, exists := g.lastValues[sensorID]; exists {

		variation := (rand.Float64() * 0.04) - 0.02

		newValue := lastValue * (1 + variation)

		noise := rand.NormFloat64() * (range_ * 0.01)
		newValue += noise

		midPoint := min + (range_ / 2)
		drift := (midPoint - newValue) * 0.01 * rand.Float64()
		newValue += drift

		if newValue < min {
			newValue = min
		}
		if newValue > max {
			newValue = max
		}

		return newValue
	}

	return min + rand.Float64()*range_
}
