package models

import "time"

type GameSession struct {
	ID        string
	GameID    string
	Player1ID string
	Player2ID string
	WinnerIdx *int
	PlayedAt  time.Time
}
