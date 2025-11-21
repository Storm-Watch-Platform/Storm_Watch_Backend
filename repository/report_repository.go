package repository

import (
	"context"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type reportRepository struct {
	db         mongo.Database
	collection string
}

// NewReportRepo tạo repo Report
func NewReportRepo(db mongo.Database, col string) domain.ReportRepository {
	return &reportRepository{db: db, collection: col}
}

// Create lưu report vào MongoDB nếu chưa tồn tại (upsert)
func (r *reportRepository) Create(ctx context.Context, rep *domain.Report) error {
	if rep.Timestamp == 0 {
		rep.Timestamp = time.Now().Unix()
	}

	if rep.ID.IsZero() {
		rep.ID = primitive.NewObjectID()
	}

	coll := r.db.Collection(r.collection)

	// Upsert: nếu _id đã tồn tại thì bỏ qua, chưa có thì insert
	filter := bson.M{"_id": rep.ID}
	update := bson.M{"$setOnInsert": rep} // chỉ insert nếu chưa tồn tại
	opts := options.Update().SetUpsert(true)

	_, err := coll.UpdateOne(ctx, filter, update, opts)
	return err
}

// UpdateAI cập nhật enrichment, nếu report chưa tồn tại sẽ tạo mới (upsert)
func (r *reportRepository) UpdateAI(ctx context.Context, reportID string, enrichment *domain.ReportEnrichment) error {
	objID, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return err
	}

	coll := r.db.Collection(r.collection)

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"enrichment": enrichment}}
	opts := options.Update().SetUpsert(true) // nếu chưa có document thì tạo mới

	_, err = coll.UpdateOne(ctx, filter, update, opts)
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
