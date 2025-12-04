package event

import "time"

type EventCreatingResponse struct {
	EventID   string    `json:"event_id"`
	CreatedAt time.Time `json:"created_at"`
}

type EventGettingResponse struct {
	ID              string          `json:"event_id"`
	Name            string          `json:"event_name"`
	Description     string          `json:"event_description"`
	EventDate       time.Time       `json:"event_date"`
	CreatedDate     time.Time       `json:"created_at"`
	TotalPlaces     int             `json:"total_places"`
	AvailablePlaces int             `json:"available_places"`
	BookedPlaces    int             `json:"booked_places"`
	Places          []PlaceResponse `json:"places"`
}

type PlaceResponse struct {
	ID       string `json:"id"`
	Row      int    `json:"row"`
	Seat     int    `json:"seat"`
	IsBooked bool   `json:"is_Booked"`
}

type EventListItemResponse struct {
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
	IsAvailablePlaces int       `json:"is_available_places"`
}
