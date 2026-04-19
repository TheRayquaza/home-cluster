package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Category struct {
	ID        bson.ObjectID  `bson:"_id,omitempty"`
	Name      string         `bson:"name"`
	ImageID   *bson.ObjectID `bson:"image_id,omitempty"`
	CreatedAt time.Time      `bson:"created_at"`
}
