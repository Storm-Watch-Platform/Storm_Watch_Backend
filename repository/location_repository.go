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

func (r *locationRepository) GetNearbyUserIDs(ctx context.Context, lat, lon, km float64) ([]string, error) {
	collection := r.database.Collection(domain.CollectionLocation)

	filter := bson.M{
		"location": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{lon, lat},
				},
				"$maxDistance": km * 1000, // km -> meters
			},
		},
	}

	// chỉ lấy field _id
	opts := options.Find().SetProjection(bson.M{"_id": 1})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ids []string
	for cursor.Next(ctx) {
		var doc struct {
			ID string `bson:"_id"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		ids = append(ids, doc.ID)
	}
	return ids, nil
}
