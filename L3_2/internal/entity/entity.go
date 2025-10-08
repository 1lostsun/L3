package entity

import "time"

const Letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Clicks struct {
	ShortID   string
	UserAgent string
	IP        string
	TimeStamp time.Time
}

type Link struct {
	ShortID     string    `json:"short_id"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	ClickCount  int       `json:"click_count"`
}

type Request struct {
	OriginalURL string `json:"url"`
}

type ClickStats struct {
	Day       time.Time `json:"day,omitempty"`        // для агрегации по дням
	UserAgent string    `json:"user_agent,omitempty"` // для агрегации по браузерам
	IP        string    `json:"ip,omitempty"`         // для агрегации по IP
	Count     int       `json:"count"`                // количество кликов
}

type Analytics struct {
	Total       uint64       `json:"total"`
	ByDay       []ClickStats `json:"by_day"`
	ByUserAgent []ClickStats `json:"by_user_agent"`
	ByIP        []ClickStats `json:"by_ip"`
}
