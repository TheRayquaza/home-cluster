package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID           bson.ObjectID  `bson:"_id,omitempty"`
	Username     string         `bson:"username"`
	Email        string         `bson:"email"`
	PasswordHash string         `bson:"password_hash"`
	Role         string         `bson:"role"` // "user" or "admin"
	PhotoID      *bson.ObjectID `bson:"photo_id,omitempty"`
	CreatedAt    time.Time      `bson:"created_at"`
}
