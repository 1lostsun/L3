package booking

import "time"

type Booking struct {
	ID          string
	EventID     string
	PlaceID     string
	Status      BookingStatus
	CreatedAt   time.Time
	PaidAt      *time.Time
	CancelledAt *time.Time
	ExpiryAt    time.Time
	UpdatedAt   time.Time
}

type BookingStatus string

const (
	PENDING   BookingStatus = "pending"
	PAID      BookingStatus = "paid"
	CANCELLED BookingStatus = "cancelled"
	EXPIRED   BookingStatus = "expired"
)
