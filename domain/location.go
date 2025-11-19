package domain

import "context"

type Location struct {
	ID        string   `bson:"_id" json:"id"` // = userID
	AccuracyM float64  `bson:"accuracy_m" json:"accuracy_m"`
	Status    string   `bson:"status" json:"status"`         // SAFE | CAUTION | DANGER | UNKNOWN
	UpdatedAt int64    `bson:"updated_at" json:"updated_at"` // timestamp (s)
	Location  GeoPoint `bson:"location" json:"location"`     // GeoJSON Point
}

type GeoPoint struct {
	Type        string     `bson:"type" json:"type"`               // "Point"
	Coordinates [2]float64 `bson:"coordinates" json:"coordinates"` // [lon, lat]
}

const CollectionLocation = "locations"

// Interface
type LocationRepository interface {
	Upsert(ctx context.Context, loc *Location) error
	GetByUserID(ctx context.Context, userID string) (*Location, error)
	GetNearbyUserIDs(ctx context.Context, lat, lon, km float64) ([]string, error)
}
