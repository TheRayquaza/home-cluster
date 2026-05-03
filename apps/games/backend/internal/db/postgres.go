package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
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
			password_hash TEXT NOT NULL,
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
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	if err := seedAdmin(db); err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}

	log.Println("db: migrations complete")
	return nil
}

func seedAdmin(db *sql.DB) error {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE username = 'admin'`).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3)`,
		"admin", string(hash), "gamemaster",
	)
	if err != nil {
		return err
	}

	log.Println("db: seeded default admin user")
	return nil
}
