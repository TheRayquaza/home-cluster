package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderItem struct {
	ArticleID   bson.ObjectID  `bson:"article_id"`
	ArticleName string         `bson:"article_name"`
	Quantity    float64        `bson:"quantity"`
	Unit        string         `bson:"unit"`
	ImageID     *bson.ObjectID `bson:"image_id,omitempty"`
}

type Order struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    bson.ObjectID `bson:"user_id"`
	Username  string        `bson:"username"`
	Date      string        `bson:"date"` // YYYY-MM-DD
	Items     []OrderItem   `bson:"items"`
	Status    string        `bson:"status"` // pending, confirmed, delivered, cancelled
	AdminNote string        `bson:"admin_note,omitempty"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}
