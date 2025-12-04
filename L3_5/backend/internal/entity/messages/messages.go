package messages

import "time"

type BookingExpiryMessage struct {
	BookingID  string    `json:"booking_id"`
	EventID    string    `json:"event_id"`
	PlaceID    string    `json:"place_id"`
	CreatedAt  time.Time `json:"created_at"`
	TTLMinutes int       `json:"ttl_minutes"`
}
