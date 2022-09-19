package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID          primitive.ObjectID `bson:"_id"`
	Title       *string            `json:"title" validate:"required,min=3,max=100"`
	Description *string            `json:"description" validate:"required,min=3,max=100"`
	Link        *string            `json:"link" validate:"required,min=3,max=100"`
	Created_at  time.Time          `json:"created_at"`
	Updated_at  time.Time          `json:"updated_at"`
	Order_id    *string            `json:"order_id"`
}
