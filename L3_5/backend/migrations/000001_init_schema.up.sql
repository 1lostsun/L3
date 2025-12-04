-- Включаем расширение для генерации UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-------------------------------------

CREATE TABLE events (
                        id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        name        VARCHAR(255) NOT NULL,
                        description TEXT,
                        event_date  TIMESTAMPTZ NOT NULL,
                        booking_ttl_minutes INT NOT NULL DEFAULT 30,
                        created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индексы для events
CREATE INDEX idx_events_event_date ON events(event_date);
CREATE INDEX idx_events_created_at ON events(created_at);
CREATE INDEX idx_events_booking_ttl ON events(booking_ttl_minutes);

-- Триггер для автообновления updated_at
CREATE TRIGGER set_timestamp_events
    BEFORE UPDATE ON events
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_timestamp();

-------------------------------------

CREATE TABLE places (
                        id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        event_id   UUID NOT NULL,
                        row_number INT NOT NULL CHECK (row_number > 0),
                        seat_number INT NOT NULL CHECK (seat_number > 0),
                        is_booked  BOOLEAN NOT NULL DEFAULT FALSE,
                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                        CONSTRAINT fk_places_event
                            FOREIGN KEY (event_id)
                                REFERENCES events(id)
                                ON DELETE CASCADE,

                        CONSTRAINT uq_places_event_row_seat
                            UNIQUE (event_id, row_number, seat_number)
);

-- Индексы для places
CREATE INDEX idx_places_event_id ON places(event_id);
CREATE INDEX idx_places_is_booked ON places(is_booked);
CREATE INDEX idx_places_event_id_is_booked ON places(event_id, is_booked);

-- Триггер для автообновления updated_at
CREATE TRIGGER set_timestamp_places
    BEFORE UPDATE ON places
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_timestamp();

-------------------------------------

CREATE TABLE bookings (
                          id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          event_id     UUID NOT NULL,
                          place_id     UUID NOT NULL,
                          status       VARCHAR(20) NOT NULL DEFAULT 'pending',
                          created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                          paid_at      TIMESTAMPTZ,
                          cancelled_at TIMESTAMPTZ,
                          expiry_at    TIMESTAMPTZ NOT NULL,
                          updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                          CONSTRAINT fk_bookings_event
                              FOREIGN KEY (event_id)
                                  REFERENCES events(id)
                                  ON DELETE CASCADE,

                          CONSTRAINT fk_bookings_place
                              FOREIGN KEY (place_id)
                                  REFERENCES places(id)
                                  ON DELETE CASCADE,

                          CONSTRAINT chk_bookings_status
                              CHECK (status IN ('pending', 'paid', 'cancelled', 'expired')),

                          CONSTRAINT chk_bookings_expiry_after_created
                              CHECK (expiry_at > created_at)
);

-- Индексы для bookings
CREATE INDEX idx_bookings_event_id ON bookings(event_id);
CREATE INDEX idx_bookings_place_id ON bookings(place_id);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_expiry_at ON bookings(expiry_at);
CREATE INDEX idx_bookings_created_at ON bookings(created_at);

-- Специальный индекс для worker'а (поиск просроченных pending броней)
CREATE INDEX idx_bookings_pending_expired
    ON bookings(status, expiry_at)
    WHERE status = 'pending';

-- Триггер для автообновления updated_at
CREATE TRIGGER set_timestamp_bookings
    BEFORE UPDATE ON bookings
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_timestamp();