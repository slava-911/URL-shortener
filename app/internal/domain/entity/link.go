package entity

import (
	"math/rand"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Link struct {
	ID           string    `json:"id"`
	FullVersion  string    `json:"full_version"`
	ShortVersion string    `json:"short_version"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	Clicked      int       `json:"clicked"`
	UserID       string    `json:"user_id"`
}

// GenerateShortVersion creates random string for short version of link
func (l *Link) GenerateShortVersion(n int) {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	l.ShortVersion = string(b)
}
