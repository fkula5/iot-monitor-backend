package services

import (
	"context"
	"log"
	"math/rand/v2"
	"sync"
	"time"

	pb_data "github.com/skni-kod/iot-monitor-backend/internal/proto/data_service"
	pb_sensor "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type IGeneratorService interface {
	Start(ctx context.Context) error
	Stop() error
	Subscribe(sensorIDs []int64) (<-chan *pb_data.ReadingUpdate, func())
}

type GeneratorService struct {
	sensorClient        pb_sensor.SensorServiceClient
	dataClient          pb_data.DataServiceClient
	generationInterval  time.Duration
	lastValues          map[int64]float64
	stopChan            chan struct{}
	generationInProcess bool
	subscribers         map[string]chan *pb_data.ReadingUpdate
	subscribersMu       sync.RWMutex
	latestReadings      map[int64]*pb_data.ReadingUpdate
	latestReadingsMu    sync.RWMutex
}

func NewGeneratorService(sensorClient pb_sensor.SensorServiceClient, dataClient pb_data.DataServiceClient) IGeneratorService {
	return &GeneratorService{
		sensorClient:        sensorClient,
		dataClient:          dataClient,
		generationInterval:  5 * time.Second,
		lastValues:          make(map[int64]float64),
		stopChan:            make(chan struct{}),
		generationInProcess: false,
		subscribers:         make(map[string]chan *pb_data.ReadingUpdate),
		latestReadings:      make(map[int64]*pb_data.ReadingUpdate),
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

	g.subscribersMu.Lock()
	for _, ch := range g.subscribers {
		close(ch)
	}
	g.subscribers = make(map[string]chan *pb_data.ReadingUpdate)
	g.subscribersMu.Unlock()

	return nil
}

func (g *GeneratorService) Subscribe(sensorIDs []int64) (<-chan *pb_data.ReadingUpdate, func()) {
	ch := make(chan *pb_data.ReadingUpdate, 100)
	id := generateSubscriberID()

	g.subscribersMu.Lock()
	g.subscribers[id] = ch
	g.subscribersMu.Unlock()

	g.latestReadingsMu.RLock()
	for _, sensorID := range sensorIDs {
		if reading, exists := g.latestReadings[sensorID]; exists {
			select {
			case ch <- reading:
			default:
			}
		}
	}
	g.latestReadingsMu.RUnlock()

	unsubscribe := func() {
		g.subscribersMu.Lock()
		if subCh, exists := g.subscribers[id]; exists {
			close(subCh)
			delete(g.subscribers, id)
		}
		g.subscribersMu.Unlock()
	}

	return ch, unsubscribe
}

func (g *GeneratorService) generateData(ctx context.Context) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	allSensors := make([]*pb_sensor.Sensor, 0)

	for userId := int64(1); userId <= 6; userId++ {
		sensors, err := g.sensorClient.ListSensors(ctxTimeout, &pb_sensor.ListSensorsRequest{
			UserId: userId,
		})
		if err != nil {
			log.Printf("Error fetching sensors for user %d: %v", userId, err)
			continue
		}
		allSensors = append(allSensors, sensors.Sensors...)
	}

	activeSensors := 0
	for _, sensor := range allSensors {
		if !sensor.Active {
			continue
		}
		activeSensors++

		sensorDetails, err := g.sensorClient.GetSensor(ctxTimeout, &pb_sensor.GetSensorRequest{Id: sensor.Id})
		if err != nil {
			log.Printf("Error fetching sensor details for sensor %d: %v", sensor.Id, err)
			continue
		}

		if sensorDetails == nil || sensorDetails.Sensor == nil {
			continue
		}

		sensorTypeId := sensorDetails.Sensor.SensorTypeId
		sensorType, err := g.sensorClient.GetSensorType(ctxTimeout, &pb_sensor.GetSensorTypeRequest{Id: sensorTypeId})
		if err != nil {
			log.Printf("Error fetching sensor type for sensor %d: %v", sensor.Id, err)
			continue
		}

		if sensorType == nil || sensorType.SensorType == nil {
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

		timestamp := time.Now()

		_, err = g.dataClient.StoreReading(ctxTimeout, &pb_data.StoreReadingRequest{
			SensorId:  sensor.Id,
			Value:     float32(generatedValue),
			Timestamp: timestamppb.New(timestamp),
		})
		if err != nil {
			log.Printf("Error storing reading for sensor %d: %v", sensor.Id, err)
		}

		update := &pb_data.ReadingUpdate{
			SensorId:   sensor.Id,
			Value:      float32(generatedValue),
			Timestamp:  timestamppb.New(timestamp),
			SensorName: sensor.Name,
			Location:   sensor.Location,
			Unit:       unit,
		}

		g.latestReadingsMu.Lock()
		g.latestReadings[sensor.Id] = update
		g.latestReadingsMu.Unlock()

		g.broadcastUpdate(update)

		log.Printf("Generated data: Sensor ID=%d, Name=%s, Location=%s, Value=%.2f %s",
			sensor.Id, sensor.Name, sensor.Location, generatedValue, unit)
	}

	log.Printf("Generated data for %d active sensors", activeSensors)
}

func (g *GeneratorService) broadcastUpdate(update *pb_data.ReadingUpdate) {
	g.subscribersMu.RLock()
	defer g.subscribersMu.RUnlock()

	for _, ch := range g.subscribers {
		select {
		case ch <- update:
		default:
		}
	}
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

func generateSubscriberID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}
	return string(b)
}
