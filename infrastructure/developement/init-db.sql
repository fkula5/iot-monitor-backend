-- Create initial sensor types
INSERT INTO sensor_types (name, model, manufacturer, description, unit, min_value, max_value, created_at)
VALUES
    ('Temperature', 'DHT22', 'Adafruit', 'Temperature sensor for ambient monitoring', '°C', -40.0, 80.0, NOW()),
    ('Humidity', 'DHT22', 'Adafruit', 'Humidity sensor for ambient monitoring', '%', 0.0, 100.0, NOW()),
    ('Pressure', 'BMP280', 'Bosch', 'Barometric pressure sensor', 'hPa', 300.0, 1100.0, NOW()),
    ('Light', 'TSL2561', 'TAOS', 'Light intensity sensor', 'lux', 0.0, 40000.0, NOW()),
    ('Motion', 'PIR HC-SR501', 'Generic', 'Passive infrared motion sensor', 'boolean', 0.0, 1.0, NOW()),
    ('CO2', 'MH-Z19', 'Winsen', 'Carbon dioxide concentration sensor', 'ppm', 0.0, 5000.0, NOW()),
    ('Soil Moisture', 'Capacitive', 'Generic', 'Soil moisture sensor', '%', 0.0, 100.0, NOW());

-- Create some initial sensors
INSERT INTO sensors (name, location, description, active, created_at, updated_at, type_sensor_type)
VALUES
    ('Living Room Temp', 'Living Room', 'Temperature sensor in the living room', true, NOW(), NOW(), 1),
    ('Kitchen Humidity', 'Kitchen', 'Humidity sensor in the kitchen', true, NOW(), NOW(), 2),
    ('Office Pressure', 'Office', 'Pressure sensor in the office', true, NOW(), NOW(), 3),
    ('Balcony Light', 'Balcony', 'Light sensor on the balcony', true, NOW(), NOW(), 4),
    ('Garage Motion', 'Garage', 'Motion sensor in the garage', true, NOW(), NOW(), 5),
    ('Bedroom CO2', 'Bedroom', 'CO2 sensor in the bedroom', true, NOW(), NOW(), 6),
    ('Garden Soil', 'Garden', 'Soil moisture sensor in the garden', true, NOW(), NOW(), 7);