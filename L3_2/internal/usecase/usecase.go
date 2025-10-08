package usecase

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"github.com/1lostsun/L3/tree/main/L3_2/internal/entity"
	"github.com/1lostsun/L3/tree/main/L3_2/internal/repo/postgres"
	"math/big"
	"os"
	"time"
)

type UseCase struct {
	db *postgres.Repository
}

func NewUseCase(db *postgres.Repository) *UseCase {
	return &UseCase{
		db: db,
	}
}

func (uc *UseCase) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	baseURL := os.Getenv("BASE_URL")

	existingLink, err := uc.db.GetByOriginalURL(ctx, originalURL)
	if err != nil && !errors.Is(sql.ErrNoRows, err) {
		return "", err
	}

	if existingLink != nil {
		return baseURL + "/s/" + existingLink.ShortID, nil
	}

	var shortID string
	for {
		shortID = RandomString(5)
		exists, err := uc.db.CheckShortIDExists(ctx, shortID)
		if err != nil {
			return "", err
		}

		if !exists {
			break
		}
	}

	link := entity.Link{
		OriginalURL: originalURL,
		ShortID:     shortID,
		CreatedAt:   time.Now(),
	}

	if err := uc.db.CreateShortURL(ctx, link); err != nil {
		return "", err
	}

	shortURL := baseURL + "/s/" + shortID

	return shortURL, nil
}

func RandomString(n int) string {
	letters := entity.Letters
	b := make([]byte, n)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[num.Int64()]
	}
	return string(b)
}

func (uc *UseCase) GetShortURL(ctx context.Context, shortID, userAgent, IP string) (*entity.Link, error) {
	link, err := uc.db.GetByShortURL(ctx, shortID)
	if err != nil {
		return nil, err
	}

	click := entity.Clicks{
		ShortID:   shortID,
		UserAgent: userAgent,
		IP:        IP,
		TimeStamp: time.Now(),
	}

	if err := uc.db.SaveClick(ctx, click); err != nil {
		return nil, err
	}

	if err := uc.db.IncrementClickCount(ctx, shortID); err != nil {
		return nil, err
	}

	return link, nil
}

func (uc *UseCase) GetAnalytics(ctx context.Context, shortID string) (*entity.Analytics, error) {
	return uc.db.GetAnalytics(ctx, shortID)
}
