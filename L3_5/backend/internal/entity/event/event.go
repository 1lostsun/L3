package event

import "time"

type Event struct {
	ID                string
	Name              string
	Description       string
	Places            []Place
	EventDate         time.Time
	BookingTTLMinutes int
	UpdatedAt         time.Time
	CreatedAt         time.Time
}

type Place struct {
	ID        string
	EventID   string
	IsBooked  bool
	Row       int
	Seat      int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EventListItem struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	EventDate         time.Time `json:"event_date"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	BookingTTLMinutes int       `json:"booking_ttl_minutes"`
	TotalPlaces       int       `json:"total_places"`
	AvailablePlaces   int       `json:"available_places"`
	BookedPlaces      int       `json:"booked_places"`
}

const (
	MinBookingTTL     = 5
	DefaultBookingTTL = 30
	MaxBookingTTL     = 10080
)
