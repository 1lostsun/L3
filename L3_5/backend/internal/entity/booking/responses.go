package booking

import "time"

type BookingResponse struct {
	BookingID  string        `json:"booking_id"`
	Status     BookingStatus `json:"status"`
	ExpiryTime time.Time     `json:"expiry_time"`
	PaidTime   *time.Time    `json:"paid_time"`
	Message    string        `json:"message"`
}
