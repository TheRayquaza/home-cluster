package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func Connect(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}

func Migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,
		`CREATE TABLE IF NOT EXISTS users (
			id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username      VARCHAR(50) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL DEFAULT '',
			role          VARCHAR(20) NOT NULL DEFAULT 'player',
			created_at    TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS daily_games (
			date    DATE PRIMARY KEY,
			game_id VARCHAR(50) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS game_sessions (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			game_id    VARCHAR(50) NOT NULL,
			player1_id UUID REFERENCES users(id),
			player2_id UUID REFERENCES users(id),
			winner_idx INT,
			played_at  TIMESTAMP DEFAULT NOW()
		)`,
		// OIDC identity — email is the unique external identifier
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS email VARCHAR(255) UNIQUE`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Println("db: migrations complete")
	return nil
}
