package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Article struct {
	ID          bson.ObjectID  `bson:"_id,omitempty"`
	Name        string         `bson:"name"`
	Description string         `bson:"description"`
	ImageID     *bson.ObjectID `bson:"image_id,omitempty"`
	Unit        string         `bson:"unit"` // kg, pièce, L, botte, etc.
	Available   bool           `bson:"available"`
	CreatedAt   time.Time      `bson:"created_at"`
	UpdatedAt   time.Time      `bson:"updated_at"`
}
