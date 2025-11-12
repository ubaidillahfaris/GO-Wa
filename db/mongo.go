package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ubaidillahfaris/whatsapp.git/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// Mongo global singleton
	Mongo     *MongoService
	mongoOnce sync.Once
)

type MongoService struct {
	Client   *mongo.Client
	Database *mongo.Database
}

type PaginatedResult struct {
	Page       int64       `json:"page"`
	Limit      int64       `json:"limit"`
	Total      int64       `json:"total"`
	perPage    int64       `json:"per_page"`
	TotalPages int64       `json:"total_pages"`
	Data       interface{} `json:"data"`
}

// InitMongoService initializes Mongo singleton
func InitMongoService() (*MongoService, error) {
	var err error

	mongoOnce.Do(func() {
		log.Println("🔗 Menghubungkan ke MongoDB...")

		username := os.Getenv("MONGO_USER")
		password := os.Getenv("MONGO_PASS")
		host := os.Getenv("MONGO_HOST")
		if host == "" {
			host = "127.0.0.1:27017"
		}
		dbName := os.Getenv("MONGO_DB")
		if dbName == "" {
			dbName = "qr_db"
		}

		uri := fmt.Sprintf("mongodb://%s:%s@%s", username, password, host)
		clientOpts := options.Client().ApplyURI(uri)
		client, e := mongo.Connect(context.Background(), clientOpts)
		if e != nil {
			err = e
			return
		}

		if e := client.Ping(context.Background(), nil); e != nil {
			err = e
			return
		}

		Mongo = &MongoService{
			Client:   client,
			Database: client.Database(dbName),
		}

		log.Println("✅ Berhasil connect ke MongoDB")
	})

	return Mongo, err
}

// =========================
// Generic CRUD Methods
// =========================

// InsertOne inserts any document into a given collection
func (m *MongoService) InsertOne(ctx context.Context, collection string, data interface{}) (*mongo.InsertOneResult, error) {
	return m.Database.Collection(collection).InsertOne(ctx, data)
}

// FindAll returns documents from a collection with optional skip/limit
func (m *MongoService) FindAll(ctx context.Context, collection string, filter bson.M, skip, limit *int64) ([]bson.M, error) {
	if filter == nil {
		filter = bson.M{}
	}

	opts := options.Find()
	if skip != nil {
		opts.SetSkip(*skip)
	}
	if limit != nil {
		opts.SetLimit(*limit)
	}

	cur, err := m.Database.Collection(collection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []bson.M
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (m *MongoService) FindAllPaginate(ctx context.Context, collection string, filter bson.M, skip, limit *int64) (*PaginatedResult, error) {

	if filter == nil {
		filter = bson.M{}
	}

	total, err := m.Database.Collection(collection).CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	opts := options.Find()
	var page int64 = 1
	var perPage int64 = 20

	if skip != nil {
		page = *skip/(*limit) + 1
		opts.SetSkip(*skip)
	}

	perPage = *limit
	opts.SetLimit(*limit)

	cur, err := m.Database.Collection(collection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []bson.M
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}

	totalPages := total / perPage
	if total%perPage != 0 {
		totalPages++
	}

	return &PaginatedResult{
		Data:       results,
		Total:      total,
		Limit:      *limit,
		Page:       page,
		perPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// FindByID returns one document by _id
func (m *MongoService) FindByID(ctx context.Context, collection string, id interface{}) (bson.M, error) {
	var result bson.M
	err := m.Database.Collection(collection).FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	return result, err
}

// Update updates document by _id
func (m *MongoService) Update(ctx context.Context, collection string, id interface{}, data bson.M) error {
	data["updated_at"] = time.Now().Unix()
	_, err := m.Database.Collection(collection).UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": data})
	return err
}

// Delete removes document by _id
func (m *MongoService) Delete(ctx context.Context, collection string, id interface{}) error {
	_, err := m.Database.Collection(collection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// =========================
// QuickResponse Helper
// =========================

// InsertQuickResponse otomatis set CreatedAt
func (m *MongoService) InsertQuickResponse(ctx context.Context, qr *models.QuickResponse) (*mongo.InsertOneResult, error) {
	now := time.Now().Unix()
	qr.CreatedAt = now
	return m.Database.Collection("quick_responses").InsertOne(ctx, qr)
}

// Optional: return quick_responses collection reference
func (m *MongoService) QuickResponseCollection() *mongo.Collection {
	return m.Database.Collection("quick_responses")
}
