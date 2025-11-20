package repositories

import (
	"context"
	"time"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/ports"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DeviceMongoRepository implements DeviceRepository using MongoDB
type DeviceMongoRepository struct {
	collection *mongo.Collection
	logger     *logger.Logger
}

// mongoDevice represents the MongoDB document structure for devices
type mongoDevice struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Owner       string             `bson:"owner"`
	Description string             `bson:"description,omitempty"`
	Status      string             `bson:"status"`
	JID         string             `bson:"jid,omitempty"`
	CreatedAt   int64              `bson:"created_at"`
	UpdatedAt   int64              `bson:"updated_at"`
}

// NewDeviceMongoRepository creates a new MongoDB device repository
func NewDeviceMongoRepository(db *mongo.Database) ports.DeviceRepository {
	collection := db.Collection("devices")

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Index on name (unique)
	_, _ = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// Index on owner
	_, _ = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "owner", Value: 1}},
	})

	return &DeviceMongoRepository{
		collection: collection,
		logger:     logger.New("DeviceRepository"),
	}
}

// Create creates a new device
func (r *DeviceMongoRepository) Create(ctx context.Context, device *domain.Device) error {
	r.logger.WithField("name", device.Name).Info("Creating device")

	doc := r.toMongoDocument(device)
	doc.CreatedAt = time.Now().Unix()
	doc.UpdatedAt = doc.CreatedAt

	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return apperrors.New(apperrors.ErrorTypeConflict, "Device with this name already exists")
		}
		r.logger.Error("Failed to create device: %v", err)
		return apperrors.NewDatabaseError("Failed to create device", err)
	}

	// Update domain entity with generated ID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		device.ID = oid.Hex()
		device.CreatedAt = time.Unix(doc.CreatedAt, 0)
		device.UpdatedAt = time.Unix(doc.UpdatedAt, 0)
	}

	r.logger.WithField("id", device.ID).Success("Device created")
	return nil
}

// FindByID retrieves a device by ID
func (r *DeviceMongoRepository) FindByID(ctx context.Context, id string) (*domain.Device, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperrors.NewValidationError("Invalid device ID format")
	}

	var doc mongoDevice
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, apperrors.NewNotFoundError("Device")
	}
	if err != nil {
		r.logger.Error("Failed to find device: %v", err)
		return nil, apperrors.NewDatabaseError("Failed to retrieve device", err)
	}

	return r.toDomainEntity(&doc), nil
}

// FindByName retrieves a device by name
func (r *DeviceMongoRepository) FindByName(ctx context.Context, name string) (*domain.Device, error) {
	var doc mongoDevice
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, apperrors.NewNotFoundError("Device")
	}
	if err != nil {
		r.logger.Error("Failed to find device by name: %v", err)
		return nil, apperrors.NewDatabaseError("Failed to retrieve device", err)
	}

	return r.toDomainEntity(&doc), nil
}

// FindAll retrieves all devices with optional filters
func (r *DeviceMongoRepository) FindAll(ctx context.Context, filter *domain.DeviceFilter, skip, limit int) ([]*domain.Device, error) {
	// Build filter
	mongoFilter := bson.M{}
	if filter != nil {
		if filter.Owner != "" {
			mongoFilter["owner"] = filter.Owner
		}
		if filter.Status != "" {
			mongoFilter["status"] = string(filter.Status)
		}
	}

	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		r.logger.Error("Failed to find devices: %v", err)
		return nil, apperrors.NewDatabaseError("Failed to retrieve devices", err)
	}
	defer cursor.Close(ctx)

	var results []*domain.Device
	for cursor.Next(ctx) {
		var doc mongoDevice
		if err := cursor.Decode(&doc); err != nil {
			r.logger.Warn("Failed to decode device: %v", err)
			continue
		}
		results = append(results, r.toDomainEntity(&doc))
	}

	if err := cursor.Err(); err != nil {
		r.logger.Error("Cursor error: %v", err)
		return nil, apperrors.NewDatabaseError("Failed to iterate devices", err)
	}

	r.logger.WithField("count", len(results)).Info("Devices retrieved")
	return results, nil
}

// Update updates a device
func (r *DeviceMongoRepository) Update(ctx context.Context, device *domain.Device) error {
	r.logger.WithField("id", device.ID).Info("Updating device")

	objectID, err := primitive.ObjectIDFromHex(device.ID)
	if err != nil {
		return apperrors.NewValidationError("Invalid device ID format")
	}

	update := bson.M{
		"$set": bson.M{
			"name":        device.Name,
			"description": device.Description,
			"status":      string(device.Status),
			"jid":         device.JID,
			"updated_at":  time.Now().Unix(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return apperrors.New(apperrors.ErrorTypeConflict, "Device with this name already exists")
		}
		r.logger.Error("Failed to update device: %v", err)
		return apperrors.NewDatabaseError("Failed to update device", err)
	}

	if result.MatchedCount == 0 {
		return apperrors.NewNotFoundError("Device")
	}

	device.UpdatedAt = time.Now()

	r.logger.WithField("id", device.ID).Success("Device updated")
	return nil
}

// Delete deletes a device (soft delete)
func (r *DeviceMongoRepository) Delete(ctx context.Context, id string) error {
	r.logger.WithField("id", id).Info("Deleting device")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return apperrors.NewValidationError("Invalid device ID format")
	}

	// Soft delete by updating status
	update := bson.M{
		"$set": bson.M{
			"status":     string(domain.DeviceStatusDeleted),
			"updated_at": time.Now().Unix(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error("Failed to delete device: %v", err)
		return apperrors.NewDatabaseError("Failed to delete device", err)
	}

	if result.MatchedCount == 0 {
		return apperrors.NewNotFoundError("Device")
	}

	r.logger.WithField("id", id).Success("Device deleted")
	return nil
}

// Count counts devices with optional filter
func (r *DeviceMongoRepository) Count(ctx context.Context, filter *domain.DeviceFilter) (int64, error) {
	// Build filter
	mongoFilter := bson.M{}
	if filter != nil {
		if filter.Owner != "" {
			mongoFilter["owner"] = filter.Owner
		}
		if filter.Status != "" {
			mongoFilter["status"] = string(filter.Status)
		}
	}

	count, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		r.logger.Error("Failed to count devices: %v", err)
		return 0, apperrors.NewDatabaseError("Failed to count devices", err)
	}

	return count, nil
}

// UpdateJID updates the JID of a device
func (r *DeviceMongoRepository) UpdateJID(ctx context.Context, id, jid string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return apperrors.NewValidationError("Invalid device ID format")
	}

	update := bson.M{
		"$set": bson.M{
			"jid":        jid,
			"updated_at": time.Now().Unix(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error("Failed to update device JID: %v", err)
		return apperrors.NewDatabaseError("Failed to update device JID", err)
	}

	if result.MatchedCount == 0 {
		return apperrors.NewNotFoundError("Device")
	}

	r.logger.WithFields(map[string]interface{}{
		"id":  id,
		"jid": jid,
	}).Success("Device JID updated")
	return nil
}

// UpdateStatus updates the status of a device
func (r *DeviceMongoRepository) UpdateStatus(ctx context.Context, id string, status domain.DeviceStatus) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return apperrors.NewValidationError("Invalid device ID format")
	}

	update := bson.M{
		"$set": bson.M{
			"status":     string(status),
			"updated_at": time.Now().Unix(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error("Failed to update device status: %v", err)
		return apperrors.NewDatabaseError("Failed to update device status", err)
	}

	if result.MatchedCount == 0 {
		return apperrors.NewNotFoundError("Device")
	}

	r.logger.WithFields(map[string]interface{}{
		"id":     id,
		"status": status,
	}).Success("Device status updated")
	return nil
}

// toMongoDocument converts domain entity to MongoDB document
func (r *DeviceMongoRepository) toMongoDocument(device *domain.Device) *mongoDevice {
	doc := &mongoDevice{
		Name:        device.Name,
		Owner:       device.Owner,
		Description: device.Description,
		Status:      string(device.Status),
		JID:         device.JID,
	}

	if device.ID != "" {
		if oid, err := primitive.ObjectIDFromHex(device.ID); err == nil {
			doc.ID = oid
		}
	}

	if !device.CreatedAt.IsZero() {
		doc.CreatedAt = device.CreatedAt.Unix()
	}
	if !device.UpdatedAt.IsZero() {
		doc.UpdatedAt = device.UpdatedAt.Unix()
	}

	return doc
}

// toDomainEntity converts MongoDB document to domain entity
func (r *DeviceMongoRepository) toDomainEntity(doc *mongoDevice) *domain.Device {
	return &domain.Device{
		ID:          doc.ID.Hex(),
		Name:        doc.Name,
		Owner:       doc.Owner,
		Description: doc.Description,
		Status:      domain.DeviceStatus(doc.Status),
		JID:         doc.JID,
		CreatedAt:   time.Unix(doc.CreatedAt, 0),
		UpdatedAt:   time.Unix(doc.UpdatedAt, 0),
	}
}
