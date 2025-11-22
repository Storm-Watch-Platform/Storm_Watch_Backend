package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	CollectionZone = "zones"
)

// Zone — khu vực cảnh báo dạng hình tròn
type Zone struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Center    GeoPoint           `bson:"center" json:"center"` // GeoJSON Point
	Radius    float64            `bson:"radius" json:"radius"` // meters
	RiskScore float64            `bson:"riskScore" json:"riskScore"`
	Label     string             `bson:"label" json:"label"`
	UpdatedAt int64              `bson:"updatedAt" json:"updatedAt"`
}

// ============================
// Repository Interface
// ============================

type ZoneRepository interface {
	Create(ctx context.Context, z *Zone) error
	FetchAll(ctx context.Context) ([]Zone, error)

	// lấy các zone trong bounding box (min/max lat/lon)
	FetchInBounds(ctx context.Context,
		minLat, minLon, maxLat, maxLon float64) ([]Zone, error)

	// tìm zone nào chứa lat/lon (lat/lon nằm trong vòng tròn)
	FetchAllByLatLon(ctx context.Context, lat, lon float64) ([]Zone, error)
	Update(ctx context.Context, z *Zone) error
}

// ============================
// Usecase Interface
// ============================

type ZoneUsecase interface {
	Create(ctx context.Context, z *Zone) error
	FetchAll(ctx context.Context) ([]Zone, error)
	FetchInBounds(ctx context.Context, minLat, minLon, maxLat, maxLon float64) ([]Zone, error)
	FetchAllByLatLon(ctx context.Context, lat, lon float64) ([]Zone, error)
	AddRiskOrCreate(ctx context.Context, lat, lon, riskIncrement, defaultRadius float64) error
}
