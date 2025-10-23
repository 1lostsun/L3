package entity

import "time"

type Comment struct {
	ID           string     `json:"id"`
	Text         string     `json:"text"`
	Date         time.Time  `json:"date"`
	Parent       *string    `json:"parent"`
	CommentsTree []*Comment `json:"commentsTree"`
}

type Request struct {
	Text   string  `json:"text"`
	Parent *string `json:"parent"`
}
