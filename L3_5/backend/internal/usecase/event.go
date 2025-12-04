package usecase

import (
	"context"
	apperrors "github.com/1lostsun/L3/internal/entity/errors"
	"github.com/1lostsun/L3/internal/entity/event"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

// validateEventRequest валидирует структуру запроса
func validateEventRequest(e *event.EventRequest, eventID string, now time.Time) error {
	if e.Seats <= 0 {
		return apperrors.ErrSeatMustBeGreaterThanZero
	}

	if e.Rows <= 0 {
		return apperrors.ErrRowMustBeGreaterThanZero
	}

	if e.BookingTTLMinutes < event.MinBookingTTL || e.BookingTTLMinutes > event.MaxBookingTTL {
		return apperrors.ErrInvalidTTL
	}

	if e.EventDate.Before(now) {
		slog.Error("event date is in the past",
			slog.String("event_id", eventID),
		)

		return apperrors.ErrEventDateInPast
	}

	return nil
}

// CreateEvent создает событие
func (uc *UseCase) CreateEvent(ctx context.Context, e *event.EventRequest) (*event.EventCreatingResponse, error) {
	eventID := uuid.New().String()
	now := time.Now()

	if err := validateEventRequest(e, eventID, now); err != nil {
		return nil, err
	}

	places := make([]event.Place, 0, e.Rows*e.Seats)

	for row := 1; row <= e.Rows; row++ {
		for seat := 1; seat <= e.Seats; seat++ {
			places = append(places, event.Place{
				ID:        uuid.New().String(),
				EventID:   eventID,
				Row:       row,
				Seat:      seat,
				CreatedAt: now,
				UpdatedAt: now,
			})
		}
	}

	newEvent := event.Event{
		ID:                eventID,
		Name:              e.Name,
		Description:       e.Description,
		Places:            places,
		EventDate:         e.EventDate,
		BookingTTLMinutes: e.BookingTTLMinutes,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	resp, err := uc.r.CreateEvent(ctx, &newEvent)
	if err != nil {
		slog.Error("failed to create event",
			slog.String("event_id", eventID),
			slog.Any("error", err),
		)
		return nil, err
	}

	return resp, nil
}

func (uc *UseCase) GetEvent(ctx context.Context, eventID string) (*event.EventGettingResponse, error) {
	if eventID == "" {
		return nil, apperrors.ErrEventIdIsRequired
	}

	domainEvent, err := uc.r.GetEvent(ctx, eventID)
	if err != nil {
		slog.Error("failed to get event",
			slog.String("event_id", eventID),
			slog.Any("error", err))
		return nil, err
	}

	availablePlaces := 0
	bookedPlaces := 0
	totalPlaces := len(domainEvent.Places)

	places := make([]event.PlaceResponse, totalPlaces)
	for i, domainPlace := range domainEvent.Places {
		places[i] = event.PlaceResponse{
			ID:       domainPlace.ID,
			Row:      domainPlace.Row,
			Seat:     domainPlace.Seat,
			IsBooked: domainPlace.IsBooked,
		}

		if domainPlace.IsBooked {
			bookedPlaces++
		} else {
			availablePlaces++
		}
	}

	eventDTO := event.EventGettingResponse{
		ID:              domainEvent.ID,
		Name:            domainEvent.Name,
		Description:     domainEvent.Description,
		EventDate:       domainEvent.EventDate,
		CreatedDate:     domainEvent.CreatedAt,
		TotalPlaces:     totalPlaces,
		AvailablePlaces: availablePlaces,
		BookedPlaces:    bookedPlaces,
		Places:          places,
	}

	return &eventDTO, nil
}

// GetAllEvents получает все доступные сейчас мероприятия
func (uc *UseCase) GetAllEvents(ctx context.Context) ([]event.EventListItemResponse, error) {
	events, err := uc.r.GetAllEvents(ctx)
	if err != nil {
		slog.Error("failed to get all events",
			slog.Any("error", err),
		)
		return nil, err
	}

	if events == nil {
		return []event.EventListItemResponse{}, nil
	}

	responses := make([]event.EventListItemResponse, len(events))
	for i, domainEvent := range events {
		responses[i] = event.EventListItemResponse{
			ID:                domainEvent.ID,
			Name:              domainEvent.Name,
			Description:       domainEvent.Description,
			EventDate:         domainEvent.EventDate,
			CreatedAt:         domainEvent.CreatedAt,
			UpdatedAt:         domainEvent.UpdatedAt,
			BookedPlaces:      domainEvent.BookedPlaces,
			TotalPlaces:       domainEvent.TotalPlaces,
			AvailablePlaces:   domainEvent.AvailablePlaces,
			BookingTTLMinutes: domainEvent.BookingTTLMinutes,
		}
	}

	return responses, nil
}

// DeleteEvent удаляет событие
func (uc *UseCase) DeleteEvent(ctx context.Context, eventID string) error {
	if eventID == "" {
		slog.Error("event_id is required")
		return apperrors.ErrEventIdIsRequired
	}

	if err := uc.r.Delete(ctx, eventID); err != nil {
		slog.Error("failed to delete event",
			slog.String("event_id", eventID),
			slog.Any("error", err),
		)
		return err
	}

	slog.Info(
		"event deleted successfully",
		slog.String("event_id", eventID),
	)
	return nil
}
