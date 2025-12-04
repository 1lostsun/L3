package errors

import "errors"

// Ошибки событий
var (
	ErrEventNotFound      = errors.New("event not found")
	ErrEventAlreadyExists = errors.New("event already exists")
	ErrEventDateInPast    = errors.New("event date is in the past")
	ErrInvalidEventDate   = errors.New("invalid event date")
	ErrInvalidTTL         = errors.New("invalid booking TTL")
	ErrEventIdIsRequired  = errors.New("event id is required")
)

// Ошибки мест
var (
	ErrPlaceNotFound             = errors.New("place not found")
	ErrPlaceAlreadyBooked        = errors.New("place already booked")
	ErrInvalidPlace              = errors.New("invalid place")
	ErrSeatMustBeGreaterThanZero = errors.New("seat number must be greater than zero")
	ErrRowMustBeGreaterThanZero  = errors.New("row number must be greater than zero")
	ErrPlaceIDIsRequired         = errors.New("place_id is required")
)

// Ошибки бронирования
var (
	ErrBookingNotFound      = errors.New("booking not found")
	ErrBookingExpired       = errors.New("booking expired")
	ErrBookingCancelled     = errors.New("booking was cancelled")
	ErrBookingAlreadyPaid   = errors.New("booking already paid")
	ErrInvalidBookingStatus = errors.New("invalid booking status")
	ErrBookingNotPending    = errors.New("booking is not in pending status")
	ErrPaymentTimedOut      = errors.New("payment time has expired")
)
