package usecase

import (
	"context"
	"github.com/1lostsun/L3/internal/entity/booking"
	apperrors "github.com/1lostsun/L3/internal/entity/errors"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

func validateBookingRequest(request booking.BookingRequest) error {
	if request.EventID == "" {
		return apperrors.ErrEventIdIsRequired
	}

	if request.PlaceID == "" {
		return apperrors.ErrPlaceIDIsRequired
	}

	return nil
}

func (uc *UseCase) CreateBooking(ctx context.Context, request booking.BookingRequest) (*booking.BookingResponse, error) {
	bookingID := uuid.New().String()
	now := time.Now()

	if err := validateBookingRequest(request); err != nil {
		return nil, err
	}

	domainEvent, err := uc.r.GetEvent(ctx, request.EventID)
	if err != nil {
		return nil, err
	}

	ttl := time.Duration(domainEvent.BookingTTLMinutes) * time.Minute
	expiry := now.Add(ttl)

	domainBooking := booking.Booking{
		ID:        bookingID,
		EventID:   request.EventID,
		PlaceID:   request.PlaceID,
		Status:    booking.PENDING,
		CreatedAt: now,
		ExpiryAt:  expiry,
		UpdatedAt: now,
	}

	slog.Info("creating booking",
		slog.String("booking_id", bookingID),
		slog.String("event_id", request.EventID),
		slog.String("place_id", request.PlaceID),
		slog.Int("ttl_minutes", domainEvent.BookingTTLMinutes),
	)

	resp, err := uc.r.CreateBooking(ctx, &domainBooking)
	if err != nil {
		slog.Error("failed to get event for booking",
			slog.String("booking_id", bookingID),
			slog.String("event_id", request.EventID),
			slog.Any("error", err),
		)
		return nil, err
	}

	if err := uc.pb.Publish(ctx, bookingID, request.EventID, request.PlaceID, ttl); err != nil {
		slog.Warn("failed to publish booking expiry message",
			slog.String("booking_id", bookingID),
			slog.Any("error", err),
		)
	}

	slog.Info("booking created successfully",
		slog.String("booking_id", bookingID),
		slog.Time("expiry_at", expiry),
	)
	return resp, nil
}

//type Booking struct {
//	ID          string
//	EventID     string
//	PlaceID     string
//	Status      BookingStatus
//	CreatedAt   time.Time
//	PaidAt      *time.Time
//	CancelledAt *time.Time
//	ExpiryAt    time.Time
//	UpdatedAt   time.Time
//}

//type BookingRequest struct {
//	EventID string `json:"event_id"`
//	PlaceID string `json:"place_id"`
//	UserID  string `json:"user_id"`
//}
