package pg

import (
	"errors"
	"fmt"
	apperrors "github.com/1lostsun/L3/internal/entity/errors"
	"github.com/1lostsun/L3/internal/entity/event"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/net/context"
)

// CreateEvent создание события
func (r *Repo) CreateEvent(ctx context.Context, e *event.Event) (*event.EventCreatingResponse, error) {
	tx, err := r.pg.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("tx begin: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	_, err = tx.Exec(ctx, `
		INSERT INTO events (id, name, description, event_date, created_at, updated_at, booking_ttl_minutes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		e.ID, e.Name, e.Description, e.EventDate, e.CreatedAt, e.UpdatedAt, e.BookingTTLMinutes,
	)

	if err != nil {
		if pgErr := handlePgError(err); pgErr != nil {
			return nil, pgErr
		}

		return nil, fmt.Errorf("failed to insert event: %w", err)
	}

	if len(e.Places) > 0 {
		batch := &pgx.Batch{}

		for _, place := range e.Places {
			batch.Queue(`
				INSERT INTO places (id, event_id, row_number, seat_number, is_booked, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				place.ID, e.ID, place.Row, place.Seat, place.IsBooked, e.CreatedAt, e.UpdatedAt,
			)
		}

		br := tx.SendBatch(ctx, batch)

		for i := 0; i < len(e.Places); i++ {
			if _, err = br.Exec(); err != nil {
				return nil, fmt.Errorf("failed to insert place %d: %w", i, err)
			}
		}

		if err := br.Close(); err != nil {
			return nil, fmt.Errorf("failed to close batch: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &event.EventCreatingResponse{
		EventID:   e.ID,
		CreatedAt: e.CreatedAt,
	}, nil
}

// GetEvent получение события
func (r *Repo) GetEvent(ctx context.Context, eventID string) (*event.Event, error) {
	var e event.Event
	err := r.pg.QueryRow(ctx, `
		SELECT
			id,
			name, 
			description,
			event_date,
			created_at,
			updated_at,
			booking_ttl_minutes,
		FROM events
		WHERE id = $1
		`,
		eventID).Scan(
		&e.ID,
		&e.Name,
		&e.Description,
		&e.EventDate,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.BookingTTLMinutes,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	rows, err := r.pg.Query(ctx, `
		SELECT
			id,
			event_id,
			row_number,
			seat_number,
			is_booked,
			created_at,
			updated_at
		FROM places
		where event_id = $1
		ORDER BY row_number, seat_number
		`, e.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to get places; eventID: %s: %w", e.ID, err)
	}

	defer rows.Close()

	places := make([]event.Place, 0, 256)
	for rows.Next() {
		var place event.Place
		err := rows.Scan(
			&place.ID,
			&place.EventID,
			&place.Row,
			&place.Seat,
			&place.IsBooked,
			&place.CreatedAt,
			&place.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan place: %w", err)
		}

		places = append(places, place)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate places rows: %w", err)
	}

	e.Places = places
	return &e, nil
}

// GetAllEvents получает все предстоящие события
func (r *Repo) GetAllEvents(ctx context.Context) ([]*event.EventListItem, error) {
	rows, err := r.pg.Query(ctx, `
		SELECT
			e.id,
			e.name,
			e.description,
			e.event_date,
			e.created_at,
			e.updated_at,
			e.booking_ttl_minutes,
			COUNT(p.id) as total_places,
			COUNT(p.id) FILTER (WHERE p.is_booked = false) as available_places,
			COUNT(p.id) FILTER (WHERE p.is_booked = true) as booked_places
		FROM events e
		LEFT JOIN places p ON p.event_id = e.id
		WHERE e.event_date > NOW()
		GROUP BY e.id
		ORDER BY e.event_date ASC
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to query list of events: %w", err)
	}

	defer rows.Close()

	events := make([]*event.EventListItem, 0, 32)
	for rows.Next() {
		var e event.EventListItem
		err := rows.Scan(
			&e.ID,
			&e.Name,
			&e.Description,
			&e.EventDate,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.TotalPlaces,
			&e.AvailablePlaces,
			&e.BookedPlaces,
			&e.BookingTTLMinutes,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}

		events = append(events, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return events, nil
}

// Delete удаляет событие
func (r *Repo) Delete(ctx context.Context, eventID string) error {
	result, err := r.pg.Exec(ctx, "DELETE FROM events WHERE id = $1", eventID)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return apperrors.ErrEventNotFound
	}

	return nil
}

func handlePgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("duplicate key: %s", pgErr.ConstraintName)
		case "23503":
			return fmt.Errorf("foreign key violation: %s", pgErr.ConstraintName)
		case "23514":
			return fmt.Errorf("check constraint violation: %s", pgErr.ConstraintName)
		case "22007":
			return fmt.Errorf("invalid date format")
		case "22001":
			return fmt.Errorf("value too long for column: %s", pgErr.ColumnName)
		}
	}

	return nil
}
