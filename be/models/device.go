// models/device.go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Device struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Owner     *string            `bson:owner" json:"owner" validate:"required"`
	Name      *string            `bson:"name" json:"name,omitempty"`
	Status    *string            `bson:"status" json:"status,omitempty"`
	CreatedAt *int64             `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt *int64             `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}
