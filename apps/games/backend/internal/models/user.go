package models

import "time"

type User struct {
	ID           string
	Username     string
	PasswordHash string
	Role         string // "player" | "gamemaster"
	CreatedAt    time.Time
}
