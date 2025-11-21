package repository

import (
	"context"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type cellRepository struct {
	db         mongo.Database
	collection string
}

func NewCellRepository(db mongo.Database, collection string) domain.CellRepository {
	return &cellRepository{
		db:         db,
		collection: collection,
	}
}

func (r *cellRepository) GetByLatLon(ctx context.Context, lat, lon float64) (*domain.Cell, error) {
	col := r.db.Collection(r.collection)
	filter := bson.M{"center": [2]float64{lon, lat}} // snapGrid đã được áp dụng
	var cell domain.Cell
	err := col.FindOne(ctx, filter).Decode(&cell)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return nil, nil
		}
		return nil, err
	}
	return &cell, nil
}

func (r *cellRepository) GetByRadius(ctx context.Context, lat, lon, radius float64) ([]domain.Cell, error) {
	col := r.db.Collection(r.collection)

	filter := bson.M{
		"center": bson.M{
			"$geoWithin": bson.M{
				"$centerSphere": []interface{}{
					[]float64{lon, lat},
					radius / 6371000.0, // convert meters to radians
				},
			},
		},
	}

	cur, err := col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var cells []domain.Cell
	if err := cur.All(ctx, &cells); err != nil {
		return nil, err
	}
	return cells, nil
}

func (r *cellRepository) Upsert(ctx context.Context, cell *domain.Cell) error {
	col := r.db.Collection(r.collection)
	filter := bson.M{"center": cell.Center}
	update := bson.M{
		"$set": bson.M{
			"riskScore": cell.RiskScore,
			"label":     cell.Label,
			"updatedAt": cell.UpdatedAt,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := col.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *cellRepository) UpsertMany(ctx context.Context, cells []*domain.Cell) error {
	for _, cell := range cells {
		if err := r.Upsert(ctx, cell); err != nil {
			return err
		}
	}
	return nil
}
