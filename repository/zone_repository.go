package repository

import (
	"context"
	"math"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type zoneRepository struct {
	db         mongo.Database
	collection string
}

func NewZoneRepository(db mongo.Database, collection string) domain.ZoneRepository {
	return &zoneRepository{
		db:         db,
		collection: collection,
	}
}

// Create zone mới
func (zr *zoneRepository) Create(ctx context.Context, z *domain.Zone) error {
	coll := zr.db.Collection(zr.collection)
	_, err := coll.InsertOne(ctx, z)
	return err
}

// FetchAll lấy tất cả zone
func (zr *zoneRepository) FetchAll(ctx context.Context) ([]domain.Zone, error) {
	coll := zr.db.Collection(zr.collection)
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var zones []domain.Zone
	err = cursor.All(ctx, &zones)
	return zones, err
}

// FetchInBounds lấy zone trong bounding box
func (zr *zoneRepository) FetchInBounds(ctx context.Context, minLat, minLon, maxLat, maxLon float64) ([]domain.Zone, error) {
	coll := zr.db.Collection(zr.collection)
	filter := bson.M{
		"center.0": bson.M{"$gte": minLon, "$lte": maxLon}, // lon
		"center.1": bson.M{"$gte": minLat, "$lte": maxLat}, // lat
	}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var zones []domain.Zone
	err = cursor.All(ctx, &zones)
	return zones, err
}

// FetchByLatLon kiểm tra point nằm trong zone nào
func (zr *zoneRepository) FetchAllByLatLon(ctx context.Context, lat, lon float64) ([]domain.Zone, error) {
	var zones []domain.Zone
	collection := zr.db.Collection(domain.CollectionZone)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &zones)
	if err != nil {
		return nil, err
	}

	var result []domain.Zone
	for _, z := range zones {
		// kiểm tra xem điểm nằm trong bán kính
		dist := distanceMeters(lat, lon, z.Center[1], z.Center[0])
		if dist <= z.Radius {
			result = append(result, z)
		}
	}

	return result, nil
}

// Hàm tính khoảng cách giữa 2 điểm (lat/lon) bằng mét
func distanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // bán kính Trái Đất
	latRad1 := lat1 * math.Pi / 180
	latRad2 := lat2 * math.Pi / 180
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(latRad1)*math.Cos(latRad2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// Update cập nhật một zone
func (zr *zoneRepository) Update(ctx context.Context, z *domain.Zone) error {
	coll := zr.db.Collection(zr.collection)
	_, err := coll.UpdateOne(
		ctx,
		bson.M{"_id": z.ID},
		bson.M{
			"$set": bson.M{
				"riskScore": z.RiskScore,
				"label":     z.Label,
				"updatedAt": z.UpdatedAt,
			},
		},
	)
	return err
}
