INSERT INTO sensor_types (name, model, manufacturer, description, unit, min_value, max_value, created_at)
VALUES 
('Temperatura Powietrza', 'DHT22', 'Aosong', 'Czujnik temperatury i wilgotności', '°C', -40, 80, NOW()),
('Wilgotność Powietrza', 'DHT22', 'Aosong', 'Czujnik temperatury i wilgotności', '%', 0, 100, NOW()),
('Ciśnienie Atmosferyczne', 'BMP280', 'Bosch', 'Precyzyjny czujnik ciśnienia', 'hPa', 300, 1100, NOW()),
('Jakość Powietrza (PM2.5)', 'PMS5003', 'Plantower', 'Czujnik pyłów zawieszonych PM2.5', 'µg/m³', 0, 500, NOW()),
('Jakość Powietrza (PM10)', 'PMS5003', 'Plantower', 'Czujnik pyłów zawieszonych PM10', 'µg/m³', 0, 500, NOW()),
('Czujnik Ruchu', 'HC-SR501', 'Generic', 'Czujnik ruchu PIR', 'bool', 0, 1, NOW()),
('Poziom Światła', 'BH1750', 'ROHM', 'Cyfrowy czujnik natężenia światła', 'lx', 0, 65535, NOW()),
('Wilgotność Gleby', 'YL-69', 'Generic', 'Analogowy czujnik wilgotności gleby', '%', 0, 100, NOW()),
('Czujnik Drzwi/Okien', 'MC-38', 'Generic', 'Czujnik magnetyczny (kontaktron)', 'bool', 0, 1, NOW()),
('Poziom Hałasu', 'MAX4466', 'Maxim Integrated', 'Moduł mikrofonu z wzmacniaczem', 'dB', 20, 100, NOW());