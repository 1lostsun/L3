package event

import "time"

type EventRequest struct {
	Name              string    `json:"event_name"`
	Description       string    `json:"description"`
	EventDate         time.Time `json:"event_date"`
	Rows              int       `json:"rows"`
	Seats             int       `json:"seats"`
	BookingTTLMinutes int       `json:"booking_ttl_minutes" binding:"min=5,max=10080"`
}
