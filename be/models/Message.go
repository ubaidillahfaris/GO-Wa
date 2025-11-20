// models/Message.go
package models

import (
	"go.mau.fi/whatsmeow/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	JID  types.JID          `bson:"jid"`
	Text string             `bson:"text"`
}
