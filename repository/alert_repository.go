package repository

import (
	"context"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
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

func (r *alertRepository) Create(c context.Context, alert *domain.Alert) error {
	coll := r.database.Collection(r.collection)
	_, err := coll.InsertOne(c, alert)
	return err
}

// fake function để sau này query bán kính 10km
func (r *alertRepository) FetchByRadius(c context.Context, lat, lng, km float64) ([]domain.Alert, error) {
	// hardcode tạm 2 alert quanh đây
	return []domain.Alert{
		{UserID: "user1", Lat: lat + 0.001, Lng: lng + 0.001, Body: "Test alert1"},
		{UserID: "user2", Lat: lat - 0.001, Lng: lng - 0.001, Body: "Test alert2"},
	}, nil
}
