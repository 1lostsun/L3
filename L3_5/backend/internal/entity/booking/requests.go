package booking

type BookingRequest struct {
	EventID string `json:"event_id"`
	PlaceID string `json:"place_id"`
	UserID  string `json:"user_id"`
}
