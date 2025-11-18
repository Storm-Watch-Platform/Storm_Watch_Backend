package repository

import (
	"context"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type locationRepository struct {
	database   mongo.Database
	collection string
}

func NewLocationRepo(db mongo.Database, collection string) domain.LocationRepository {
	return &locationRepository{
		database:   db,
		collection: collection,
	}
}

func (r *locationRepository) Upsert(ctx context.Context, loc *domain.Location) error {
	coll := r.database.Collection(r.collection)

	_, err := coll.UpdateOne(
		ctx,
		bson.M{"_id": loc.ID}, // _id = userID
		bson.M{
			"$set": bson.M{
				"accuracy_m": loc.AccuracyM,
				"status":     loc.Status,
				"updated_at": loc.UpdatedAt,
				"location": bson.M{ // GeoJSON Point
					"type":        loc.Location.Type,
					"coordinates": loc.Location.Coordinates, // [lon, lat]
				},
			},
		},
		options.Update().SetUpsert(true),
	)

	return err
}

func (r *locationRepository) GetByUserID(ctx context.Context, userID string) (*domain.Location, error) {
	coll := r.database.Collection(r.collection)

	var loc domain.Location
	err := coll.FindOne(ctx, bson.M{"_id": userID}).Decode(&loc)

	return &loc, err
}
