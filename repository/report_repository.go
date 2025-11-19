package repository

import (
	"context"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type reportRepository struct {
	db         mongo.Database
	collection string
}

// NewReportRepo tạo repo Report
func NewReportRepo(db mongo.Database, col string) domain.ReportRepository {
	return &reportRepository{db: db, collection: col}
}

// Create lưu report vào MongoDB
func (r *reportRepository) Create(ctx context.Context, rep *domain.Report) error {
	coll := r.db.Collection(r.collection)

	if rep.Timestamp == 0 {
		rep.Timestamp = time.Now().Unix()
	}

	// rep.Image là Base64 string, lưu trực tiếp vào DB
	_, err := coll.InsertOne(ctx, rep)
	return err
}

func (r *reportRepository) UpdateAI(ctx context.Context, reportID string, enrichment *domain.ReportEnrichment) error {
	coll := r.db.Collection(r.collection)

	objID, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return err
	}

	_, err = coll.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"enrichment": enrichment}},
	)
	return err
}

// ---------- repository/report_repository.go ----------
func (r *reportRepository) GetNearbyReports(ctx context.Context, lat, lon, km float64) ([]*domain.Report, error) {
	collection := r.db.Collection(r.collection)

	filter := bson.M{
		"location": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{lon, lat}, // [lon, lat]
				},
				"$maxDistance": km * 1000, // km -> meters
			},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reports []*domain.Report
	for cursor.Next(ctx) {
		var rep domain.Report
		if err := cursor.Decode(&rep); err != nil {
			return nil, err
		}
		reports = append(reports, &rep)
	}
	return reports, nil
}
