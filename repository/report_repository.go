// ---------- repository/report_repository.go ----------
package repository

import (
	"context"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type reportRepository struct {
	database   mongo.Database
	collection string
}

func NewReportRepo(db mongo.Database, collection string) domain.ReportRepository {
	return &reportRepository{
		database:   db,
		collection: collection,
	}
}

// Create lưu report vào MongoDB
func (r *reportRepository) Create(ctx context.Context, report *domain.Report) error {
	if ctx == nil {
		ctx = context.Background()
	}
	coll := r.database.Collection(r.collection)
	_, err := coll.InsertOne(ctx, report)
	return err
}

// FetchByGroup: lấy danh sách report user trong cùng group
// tạm thời hardcode danh sách user, sau này sẽ query MongoDB
func (r *reportRepository) FetchByGroup(ctx context.Context, groupID string) ([]domain.ReportUser, error) {
	// hardcode tạm, sẽ thay bằng query DB
	users := []domain.ReportUser{
		{UserID: "userA"},
		{UserID: "userB"},
	}
	return users, nil
}

func (r *reportRepository) UpdateEnrichment(ctx context.Context, id string, enrich *domain.ReportEnrichment) error {
	coll := r.database.Collection(r.collection)

	_, err := coll.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"enrichment": enrich,
		}},
	)

	return err
}
