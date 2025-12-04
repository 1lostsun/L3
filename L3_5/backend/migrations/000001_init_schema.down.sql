-- Удаляем таблицы в обратном порядке
DROP TRIGGER IF EXISTS set_timestamp_bookings ON bookings;
DROP TABLE IF EXISTS bookings CASCADE;

DROP TRIGGER IF EXISTS set_timestamp_places ON places;
DROP TABLE IF EXISTS places CASCADE;

DROP TRIGGER IF EXISTS set_timestamp_events ON events;
DROP TABLE IF EXISTS events CASCADE;

-- Удаляем функцию триггера
DROP FUNCTION IF EXISTS trigger_set_timestamp();

-- Удаляем расширение UUID
DROP EXTENSION IF EXISTS "uuid-ossp";
