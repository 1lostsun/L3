package entity

import "time"

type NotificationStatus string
type Retry int

const (
	StatusScheduled NotificationStatus = "scheduled"
	StatusSent      NotificationStatus = "sent"
	StatusCancelled NotificationStatus = "cancelled"
	StatusFailed    NotificationStatus = "failed"

	RetriesCount Retry = 5

	TTL = time.Hour * 24
)

type Notification struct {
	ID      uint64             `json:"id"`
	Message string             `json:"message"`
	Date    time.Time          `json:"date"`
	Status  NotificationStatus `json:"status"`
	Retries Retry
}

type Response struct {
	Data interface{} `json:"data,omitempty"`
	Err  string      `json:"error,omitempty"`
}
