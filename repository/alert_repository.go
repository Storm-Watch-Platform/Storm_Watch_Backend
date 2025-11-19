package repository

import (
	"context"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type alertRepository struct {
	database   mongo.Database
	collection string
}

func NewAlertRepo(db mongo.Database, collection string) domain.AlertRepository {
	return &alertRepository{
		database:   db,
		collection: collection,
	}
}

// Create lưu alert và tính ExpiresAt dựa trên TTLMin nếu chưa có
func (r *alertRepository) Create(ctx context.Context, alert *domain.Alert) error {
	if alert.ID.IsZero() {
		alert.ID = primitive.NewObjectID()
	}
	if alert.TTLMin > 0 && alert.ExpiresAt.IsZero() {
		alert.ExpiresAt = time.Now().Add(time.Duration(alert.TTLMin) * time.Minute)
	}
	if alert.Visibility == "" {
		alert.Visibility = "PUBLIC"
	}
	if alert.Status == "" {
		alert.Status = "RAISED"
	}

	coll := r.database.Collection(r.collection)
	_, err := coll.InsertOne(ctx, alert)
	return err
}

// UpdateStatus cho Resolve
func (r *alertRepository) UpdateStatus(ctx context.Context, alertID string, status string) error {
	coll := r.database.Collection(r.collection)
	id, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		return err
	}

	_, err = coll.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": status}},
	)
	return err
}

// FetchByID
func (r *alertRepository) FetchByID(ctx context.Context, alertID string) (*domain.Alert, error) {
	coll := r.database.Collection(r.collection)
	id, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		return nil, err
	}

	var alert domain.Alert
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&alert)
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// FetchByRadius: fake tạm, trả về những alert/user gần
func (r *alertRepository) FetchByRadius(ctx context.Context, lat, lng, km float64) ([]domain.Alert, error) {
	// hardcode tạm
	return []domain.Alert{
		{
			UserID: "user1",
			Location: domain.GeoPoint{
				Type:        "Point",
				Coordinates: [2]float64{lng + 0.001, lat + 0.001},
			},
			Body: "Test alert1",
		},
		{
			UserID: "user2",
			Location: domain.GeoPoint{
				Type:        "Point",
				Coordinates: [2]float64{lng - 0.001, lat - 0.001},
			},
			Body: "Test alert2",
		},
	}, nil
}
