package repositories

import (
	"context"
	"time"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// APIKeyMongoRepository implements the APIKeyRepository interface using MongoDB
type APIKeyMongoRepository struct {
	collection *mongo.Collection
	logger     *logger.Logger
}

// NewAPIKeyMongoRepository creates a new instance of APIKeyMongoRepository
func NewAPIKeyMongoRepository(db *mongo.Database, log *logger.Logger) (*APIKeyMongoRepository, error) {
	collection := db.Collection("api_keys")

	repo := &APIKeyMongoRepository{
		collection: collection,
		logger:     log.WithPrefix("APIKeyRepo"),
	}

	// Create indexes
	if err := repo.createIndexes(context.Background()); err != nil {
		return nil, errors.Wrap(err, errors.ErrTypeDatabase, "failed to create indexes")
	}

	return repo, nil
}

// createIndexes creates necessary indexes for the API keys collection
func (r *APIKeyMongoRepository) createIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "key", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "owner", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "expires_at", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}

	r.logger.Info("Created indexes for api_keys collection")
	return nil
}

// Create creates a new API key
func (r *APIKeyMongoRepository) Create(ctx context.Context, apiKey *domain.APIKey) error {
	if apiKey.ID == "" {
		apiKey.ID = primitive.NewObjectID().Hex()
	}

	apiKey.CreatedAt = time.Now()
	apiKey.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, apiKey)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			r.logger.Error("Duplicate API key", err)
			return errors.New(errors.ErrTypeValidation, "API key already exists")
		}
		r.logger.Error("Failed to create API key", err)
		return errors.Wrap(err, errors.ErrTypeDatabase, "failed to create API key")
	}

	r.logger.Info("API key created", logger.Fields{"id": apiKey.ID, "owner": apiKey.Owner})
	return nil
}

// GetByID retrieves an API key by its ID
func (r *APIKeyMongoRepository) GetByID(ctx context.Context, id string) (*domain.APIKey, error) {
	var apiKey domain.APIKey

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&apiKey)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New(errors.ErrTypeNotFound, "API key not found")
		}
		r.logger.Error("Failed to get API key by ID", err, logger.Fields{"id": id})
		return nil, errors.Wrap(err, errors.ErrTypeDatabase, "failed to get API key")
	}

	return &apiKey, nil
}

// GetByKey retrieves an API key by its key value
func (r *APIKeyMongoRepository) GetByKey(ctx context.Context, key string) (*domain.APIKey, error) {
	var apiKey domain.APIKey

	err := r.collection.FindOne(ctx, bson.M{"key": key}).Decode(&apiKey)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New(errors.ErrTypeNotFound, "API key not found")
		}
		r.logger.Error("Failed to get API key by key", err)
		return nil, errors.Wrap(err, errors.ErrTypeDatabase, "failed to get API key")
	}

	return &apiKey, nil
}

// List retrieves all API keys for a specific owner with pagination
func (r *APIKeyMongoRepository) List(ctx context.Context, owner string, limit, offset int) ([]*domain.APIKey, int64, error) {
	filter := bson.M{"owner": owner}

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to count API keys", err, logger.Fields{"owner": owner})
		return nil, 0, errors.Wrap(err, errors.ErrTypeDatabase, "failed to count API keys")
	}

	// Set default limit if not provided
	if limit <= 0 {
		limit = 50
	}

	// Find documents with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("Failed to list API keys", err, logger.Fields{"owner": owner})
		return nil, 0, errors.Wrap(err, errors.ErrTypeDatabase, "failed to list API keys")
	}
	defer cursor.Close(ctx)

	var apiKeys []*domain.APIKey
	if err := cursor.All(ctx, &apiKeys); err != nil {
		r.logger.Error("Failed to decode API keys", err)
		return nil, 0, errors.Wrap(err, errors.ErrTypeDatabase, "failed to decode API keys")
	}

	r.logger.Debug("Listed API keys", logger.Fields{"owner": owner, "count": len(apiKeys), "total": total})
	return apiKeys, total, nil
}

// Update updates an existing API key
func (r *APIKeyMongoRepository) Update(ctx context.Context, apiKey *domain.APIKey) error {
	apiKey.UpdatedAt = time.Now()

	filter := bson.M{"_id": apiKey.ID}
	update := bson.M{"$set": apiKey}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update API key", err, logger.Fields{"id": apiKey.ID})
		return errors.Wrap(err, errors.ErrTypeDatabase, "failed to update API key")
	}

	if result.MatchedCount == 0 {
		return errors.New(errors.ErrTypeNotFound, "API key not found")
	}

	r.logger.Info("API key updated", logger.Fields{"id": apiKey.ID})
	return nil
}

// Delete deletes an API key by ID
func (r *APIKeyMongoRepository) Delete(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to delete API key", err, logger.Fields{"id": id})
		return errors.Wrap(err, errors.ErrTypeDatabase, "failed to delete API key")
	}

	if result.DeletedCount == 0 {
		return errors.New(errors.ErrTypeNotFound, "API key not found")
	}

	r.logger.Info("API key deleted", logger.Fields{"id": id})
	return nil
}

// UpdateLastUsed updates the last used timestamp for an API key
func (r *APIKeyMongoRepository) UpdateLastUsed(ctx context.Context, key string) error {
	now := time.Now()
	filter := bson.M{"key": key}
	update := bson.M{
		"$set": bson.M{
			"last_used_at": now,
			"updated_at":   now,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update last used timestamp", err)
		return errors.Wrap(err, errors.ErrTypeDatabase, "failed to update last used timestamp")
	}

	if result.MatchedCount == 0 {
		return errors.New(errors.ErrTypeNotFound, "API key not found")
	}

	return nil
}

// CleanupExpired marks expired API keys as expired
func (r *APIKeyMongoRepository) CleanupExpired(ctx context.Context) (int64, error) {
	now := time.Now()
	filter := bson.M{
		"status": domain.APIKeyStatusActive,
		"expires_at": bson.M{
			"$ne":  nil,
			"$lte": now,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"status":     domain.APIKeyStatusExpired,
			"updated_at": now,
		},
	}

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to cleanup expired API keys", err)
		return 0, errors.Wrap(err, errors.ErrTypeDatabase, "failed to cleanup expired API keys")
	}

	if result.ModifiedCount > 0 {
		r.logger.Info("Cleaned up expired API keys", logger.Fields{"count": result.ModifiedCount})
	}

	return result.ModifiedCount, nil
}
