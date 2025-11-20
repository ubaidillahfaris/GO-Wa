package services

import (
	"context"
	"fmt"
	"log"

	"github.com/ubaidillahfaris/whatsapp.git/db"
	"github.com/ubaidillahfaris/whatsapp.git/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeviceService struct {
}

func NewDeviceService() *DeviceService {
	return &DeviceService{}
}

func (d *DeviceService) Find(ctx context.Context, id string) (models.Device, error) {
	log.Println("ID :\n", id)

	var result models.Device
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return models.Device{}, fmt.Errorf("invalid ObjectID: %w", err)
	}

	err = db.Mongo.Database.Collection("devices").FindOne(ctx, bson.M{"_id": oid}).Decode(&result)
	if err != nil {
		return models.Device{}, err
	}

	return result, nil

}
