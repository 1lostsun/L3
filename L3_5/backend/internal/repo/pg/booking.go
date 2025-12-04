package pg

import (
	"context"
	"errors"
	"fmt"
	"github.com/1lostsun/L3/internal/entity/booking"
	apperrors "github.com/1lostsun/L3/internal/entity/errors"
	"github.com/jackc/pgx/v5"
	"time"
)

// CreateBooking создает бронь на мероприятие
func (r *Repo) CreateBooking(ctx context.Context, book *booking.Booking) (*booking.BookingResponse, error) {
	tx, err := r.pg.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	var isBooked bool
	err = tx.QueryRow(ctx,
		`
		SELECT is_booked
		FROM places WHERE id = $1 AND event_id = $2 
		FOR UPDATE -- блокировка транзакции
		`,
		book.PlaceID, book.EventID).Scan(
		&isBooked)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrPlaceNotFound
		}
		return nil, fmt.Errorf("error getting isBooked: %w", err)
	}

	if isBooked {
		return nil, fmt.Errorf("place already booked")
	}

	_, err = tx.Exec(ctx,
		`
		INSERT INTO bookings (
			id, event_id, place_id, status, created_at, expiry_at	                      
		) VALUES ($1, $2, $3, $4, NOW(), $5)
		`,
		book.ID,
		book.EventID,
		book.PlaceID,
		book.Status,
		book.ExpiryAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE places
		SET is_booked = true, updated_at = NOW()
		WHERE id = $1
		`, book.PlaceID)

	if err != nil {
		return nil, fmt.Errorf("failed to update place: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &booking.BookingResponse{BookingID: book.ID,
		Status:     book.Status,
		ExpiryTime: book.ExpiryAt,
		Message:    "Booking created successfully",
	}, nil
}

// PayBooking оплата брони
func (r *Repo) PayBooking(ctx context.Context, bookingID string) (*booking.BookingResponse, error) {
	tx, err := r.pg.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}

	defer tx.Rollback(ctx)

	var b booking.Booking
	err = tx.QueryRow(ctx, `
		SELECT id, event_id, place_id, status, created_at, expiry_at
		FROM bookings
		WHERE id = $1
		FOR UPDATE 
		`, bookingID).Scan(
		&b.ID,
		&b.EventID,
		&b.PlaceID,
		&b.Status,
		&b.CreatedAt,
		&b.ExpiryAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("b not found")
		}

		return nil, fmt.Errorf("error getting b: %v", err)
	}

	if b.Status != booking.PENDING {
		if b.Status == booking.PAID {
			return nil, fmt.Errorf("b already paid")
		}

		if b.Status == booking.CANCELLED {
			return nil, fmt.Errorf("b was cancelled")
		}

		if b.Status == booking.EXPIRED {
			return nil, fmt.Errorf("b was expired")
		}

		return nil, fmt.Errorf("invalid b status: %s", b.Status)
	}

	_, err = tx.Exec(ctx, `
		UPDATE bookings
		SET status = $1,
		    paid_at = NOW()
		WHERE id = $2
		`, booking.PAID, b.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to update b: %v", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	now := time.Now()
	return &booking.BookingResponse{
		BookingID:  b.ID,
		Status:     booking.PAID,
		ExpiryTime: b.ExpiryAt,
		PaidTime:   &now,
		Message:    "Booking paid successfully",
	}, nil
}

// CancelBooking закрывает бронирование места на событии
func (r *Repo) CancelBooking(ctx context.Context, bookingID string) (bool, error) {
	tx, err := r.pg.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("error starting transaction: %v", err)
	}

	defer tx.Rollback(ctx)

	var status string
	var placeID string

	err = tx.QueryRow(ctx, `
		SELECT
		    status,
			place_id
		FROM bookings
		WHERE id = $1
		FOR UPDATE
	`, bookingID).Scan(&placeID, &status)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("booking not found")
		}

		return false, fmt.Errorf("error getting booking: %v", err)
	}

	if status != string(booking.PENDING) {
		return false, nil
	}

	_, err = tx.Exec(ctx, `
		UPDATE bookings
		SET status = $1,
			cancelled_at = NOW(),
			updated_at = NOW()
		WHERE id = $2
	`, booking.EXPIRED, bookingID)

	if err != nil {
		return false, fmt.Errorf("failed to update booking: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE places
		SET is_booked = false,
		    updated_at = NOW()
		WHERE id = $1
	`, placeID)

	if err != nil {
		return false, fmt.Errorf("failed to free place: %v", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return true, nil
}
