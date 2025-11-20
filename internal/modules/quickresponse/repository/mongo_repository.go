package repository

import (
	"context"
	"time"

	"github.com/ubaidillahfaris/whatsapp.git/internal/modules/quickresponse/domain"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoRepository implements QuickResponseRepository using MongoDB
type MongoRepository struct {
	collection *mongo.Collection
	logger     *logger.Logger
}

// mongoQuickResponse represents the MongoDB document structure
type mongoQuickResponse struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty"`
	Petugas                mongoOfficer       `bson:"petugas"`
	IdentifikasiKegiatanQR mongoActivity      `bson:"identifikasi_kegiatan_qr"`
	OutputKegiatanQR       mongoOutput        `bson:"output_kegiatan_qr"`
	CreatedAt              int64              `bson:"created_at"`
}

type mongoOfficer struct {
	Nama        string `bson:"nama"`
	Jabatan     string `bson:"jabatan"`
	DiPenugasan string `bson:"di_penugasan"`
}

type mongoActivity struct {
	MetodePenugasan    string `bson:"metode_penugasan"`
	KegiatanQR         string `bson:"kegiatan_qr"`
	DIQR               string `bson:"di_qr"`
	SaluranQR          string `bson:"saluran_qr"`
	RuasBangunanQR     string `bson:"ruas_bangunan_qr"`
	DesaKecamatanKabQR string `bson:"desa_kecamatan_kab_qr"`
	UPTPSDAWS          string `bson:"upt_psda_ws"`
}

type mongoOutput struct {
	LuasAreaKegiatan  string `bson:"luas_area_kegiatan"`
	PanjangSaluran    string `bson:"panjang_saluran"`
	MenutupBocoran    string `bson:"menutup_bocoran"`
	AngkatSedimen     string `bson:"angkat_sedimen"`
	PembersihanSampah string `bson:"pembersihan_sampah"`
	AngkatPotongPohon string `bson:"angkat_potong_pohon"`
}

// NewMongoRepository creates a new MongoDB repository for QuickResponse
func NewMongoRepository(db *mongo.Database) domain.QuickResponseRepository {
	return &MongoRepository{
		collection: db.Collection("quick_responses"),
		logger:     logger.New("QuickResponseRepository"),
	}
}

// Save saves a quick response report
func (r *MongoRepository) Save(qr *domain.QuickResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert domain to mongo document
	doc := r.toMongoDocument(qr)

	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		r.logger.Error("Failed to insert quick response: %v", err)
		return apperrors.NewDatabaseError("Failed to save quick response", err)
	}

	// Update domain entity with generated ID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		qr.ID = oid.Hex()
	}

	r.logger.WithField("id", qr.ID).Success("Quick response saved")
	return nil
}

// FindByID retrieves a quick response by ID
func (r *MongoRepository) FindByID(id string) (*domain.QuickResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperrors.NewValidationError("Invalid ID format")
	}

	var doc mongoQuickResponse
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, apperrors.NewNotFoundError("Quick response")
	}
	if err != nil {
		r.logger.Error("Failed to find quick response: %v", err)
		return nil, apperrors.NewDatabaseError("Failed to retrieve quick response", err)
	}

	return r.toDomainEntity(&doc), nil
}

// FindAll retrieves all quick responses with pagination
func (r *MongoRepository) FindAll(skip, limit int) ([]*domain.QuickResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		r.logger.Error("Failed to find quick responses: %v", err)
		return nil, apperrors.NewDatabaseError("Failed to retrieve quick responses", err)
	}
	defer cursor.Close(ctx)

	var results []*domain.QuickResponse
	for cursor.Next(ctx) {
		var doc mongoQuickResponse
		if err := cursor.Decode(&doc); err != nil {
			r.logger.Warn("Failed to decode document: %v", err)
			continue
		}
		results = append(results, r.toDomainEntity(&doc))
	}

	if err := cursor.Err(); err != nil {
		r.logger.Error("Cursor error: %v", err)
		return nil, apperrors.NewDatabaseError("Failed to iterate quick responses", err)
	}

	r.logger.WithField("count", len(results)).Info("Quick responses retrieved")
	return results, nil
}

// Delete removes a quick response
func (r *MongoRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return apperrors.NewValidationError("Invalid ID format")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		r.logger.Error("Failed to delete quick response: %v", err)
		return apperrors.NewDatabaseError("Failed to delete quick response", err)
	}

	if result.DeletedCount == 0 {
		return apperrors.NewNotFoundError("Quick response")
	}

	r.logger.WithField("id", id).Success("Quick response deleted")
	return nil
}

// Count counts total quick responses
func (r *MongoRepository) Count() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		r.logger.Error("Failed to count quick responses: %v", err)
		return 0, apperrors.NewDatabaseError("Failed to count quick responses", err)
	}

	return count, nil
}

// toMongoDocument converts domain entity to MongoDB document
func (r *MongoRepository) toMongoDocument(qr *domain.QuickResponse) *mongoQuickResponse {
	return &mongoQuickResponse{
		Petugas: mongoOfficer{
			Nama:        qr.Officer.Name,
			Jabatan:     qr.Officer.Position,
			DiPenugasan: qr.Officer.Assignment,
		},
		IdentifikasiKegiatanQR: mongoActivity{
			MetodePenugasan:    qr.Activity.Method,
			KegiatanQR:         qr.Activity.ActivityType,
			DIQR:               qr.Activity.IrrigationDI,
			SaluranQR:          qr.Activity.Channel,
			RuasBangunanQR:     qr.Activity.BuildingRoute,
			DesaKecamatanKabQR: qr.Activity.Location,
			UPTPSDAWS:          qr.Activity.WatershedUnit,
		},
		OutputKegiatanQR: mongoOutput{
			LuasAreaKegiatan:  qr.Output.AreaSize,
			PanjangSaluran:    qr.Output.ChannelLength,
			MenutupBocoran:    qr.Output.LeaksClosed,
			AngkatSedimen:     qr.Output.SedimentRemoved,
			PembersihanSampah: qr.Output.TrashCleared,
			AngkatPotongPohon: qr.Output.TreeCutRemoved,
		},
		CreatedAt: qr.CreatedAt.Unix(),
	}
}

// toDomainEntity converts MongoDB document to domain entity
func (r *MongoRepository) toDomainEntity(doc *mongoQuickResponse) *domain.QuickResponse {
	return &domain.QuickResponse{
		ID: doc.ID.Hex(),
		Officer: domain.OfficerInfo{
			Name:       doc.Petugas.Nama,
			Position:   doc.Petugas.Jabatan,
			Assignment: doc.Petugas.DiPenugasan,
		},
		Activity: domain.ActivityInfo{
			Method:        doc.IdentifikasiKegiatanQR.MetodePenugasan,
			ActivityType:  doc.IdentifikasiKegiatanQR.KegiatanQR,
			IrrigationDI:  doc.IdentifikasiKegiatanQR.DIQR,
			Channel:       doc.IdentifikasiKegiatanQR.SaluranQR,
			BuildingRoute: doc.IdentifikasiKegiatanQR.RuasBangunanQR,
			Location:      doc.IdentifikasiKegiatanQR.DesaKecamatanKabQR,
			WatershedUnit: doc.IdentifikasiKegiatanQR.UPTPSDAWS,
		},
		Output: domain.OutputInfo{
			AreaSize:        doc.OutputKegiatanQR.LuasAreaKegiatan,
			ChannelLength:   doc.OutputKegiatanQR.PanjangSaluran,
			LeaksClosed:     doc.OutputKegiatanQR.MenutupBocoran,
			SedimentRemoved: doc.OutputKegiatanQR.AngkatSedimen,
			TrashCleared:    doc.OutputKegiatanQR.PembersihanSampah,
			TreeCutRemoved:  doc.OutputKegiatanQR.AngkatPotongPohon,
		},
		CreatedAt: time.Unix(doc.CreatedAt, 0),
	}
}
