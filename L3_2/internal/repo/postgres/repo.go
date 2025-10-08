package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/1lostsun/L3/tree/main/L3_2/internal/entity"
	"github.com/wb-go/wbf/dbpg"
	"time"
)

type Repository struct {
	db *dbpg.DB
}

func NewRepository(db *dbpg.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateShortURL(ctx context.Context, link entity.Link) error {
	clickCount := 0
	query := "insert into links (short_id, original_url, created_at, click_count) values ($1, $2, $3, $4)"
	_, err := r.db.ExecContext(ctx, query, link.ShortID, link.OriginalURL, time.Now(), clickCount)
	if err != nil {
		return fmt.Errorf("error while inserting link: %w", err)
	}

	return nil
}

func (r *Repository) GetByOriginalURL(ctx context.Context, originalURL string) (*entity.Link, error) {
	query := "select * from links where original_url = $1"
	var link entity.Link
	if err := r.db.QueryRowContext(ctx, query, originalURL).Scan(
		&link.ShortID,
		&link.OriginalURL,
		&link.CreatedAt,
		&link.ClickCount,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &link, nil
}

func (r *Repository) CheckShortIDExists(ctx context.Context, shortID string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM links WHERE short_id = $1)"
	var exists bool
	err := r.db.QueryRowContext(ctx, query, shortID).Scan(&exists)
	return exists, err
}

func (r *Repository) GetByShortURL(ctx context.Context, shortID string) (*entity.Link, error) {
	var link entity.Link

	query := "select * from links where short_id = $1"
	err := r.db.QueryRowContext(ctx, query, shortID).Scan(&link.ShortID, &link.OriginalURL, &link.CreatedAt, &link.ClickCount)
	if err != nil {
		return nil, fmt.Errorf("error while getting link: %w", err)
	}

	return &link, nil
}

func (r *Repository) SaveClick(ctx context.Context, click entity.Clicks) error {
	query := "insert into clicks (short_id, user_agent, ip, timestamp) values ($1, $2, $3, $4)"

	_, err := r.db.ExecContext(ctx, query, click.ShortID, click.UserAgent, click.IP, click.TimeStamp)
	if err != nil {
		return fmt.Errorf("error while inserting click: %w", err)
	}
	return nil
}

func (r *Repository) IncrementClickCount(ctx context.Context, shortID string) error {
	query := "update links set click_count = click_count + 1 where short_id = $1"
	_, err := r.db.ExecContext(ctx, query, shortID)
	if err != nil {
		return fmt.Errorf("error while updating click count: %w", err)
	}

	return nil
}

func (r *Repository) GetAnalytics(ctx context.Context, shortID string) (*entity.Analytics, error) {
	stats := &entity.Analytics{}

	err := r.db.QueryRowContext(ctx,
		"SELECT click_count FROM links WHERE short_id = $1",
		shortID,
	).Scan(&stats.Total)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			stats.Total = 0
			return stats, nil
		}
		return nil, fmt.Errorf("error while getting analytics: %w", err)
	}

	rows, _ := r.db.QueryContext(ctx, `
        select DATE(timestamp), COUNT(*) 
        from clicks 
        where short_id=$1 
        group by DATE(timestamp) 
        order by DATE(timestamp)
    `, shortID)
	defer rows.Close()
	for rows.Next() {
		var s entity.ClickStats
		if err := rows.Scan(&s.Day, &s.Count); err != nil {
			return nil, err
		}
		stats.ByDay = append(stats.ByDay, s)
	}

	rowsUA, _ := r.db.QueryContext(ctx, `
        select user_agent, COUNT(*) 
        from clicks 
        where short_id=$1 
        group by user_agent 
        order by count DESC
    `, shortID)
	defer rowsUA.Close()
	for rowsUA.Next() {
		var s entity.ClickStats
		if err := rowsUA.Scan(&s.UserAgent, &s.Count); err != nil {
			return nil, err
		}
		stats.ByUserAgent = append(stats.ByUserAgent, s)
	}

	rowsIP, _ := r.db.QueryContext(ctx, `
        select ip, COUNT(*) 
        from clicks 
        where short_id=$1 
        group by ip 
        order by count DESC
    `, shortID)
	defer rowsIP.Close()
	for rowsIP.Next() {
		var s entity.ClickStats
		if err := rowsIP.Scan(&s.IP, &s.Count); err != nil {
			return nil, err
		}
		stats.ByIP = append(stats.ByIP, s)
	}

	return stats, nil
}
