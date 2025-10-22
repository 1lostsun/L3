package entity

import "time"

type Comment struct {
	ID           string
	Text         string
	Date         time.Time
	Parent       *string
	CommentsTree []*Comment
}

type Request struct {
	Text   string  `json:"text"`
	Parent *string `json:"parent"`
}
