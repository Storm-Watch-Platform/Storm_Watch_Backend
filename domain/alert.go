package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionAlert = "alerts"

type Alert struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"alertId"`
	UserID      string             `bson:"userID"`
	Location    GeoPoint           `bson:"location" json:"location"`
	Body        string             `bson:"body"`
	RadiusM     float64            `bson:"radius_m"`
	TTLMin      int                `bson:"ttl_min"`
	ExpiresAt   time.Time          `bson:"expires_at"`
	Visibility  string             `bson:"visibility"` // userIDs nhìn thấy alert
	Status      string             `bson:"status"`
	UserName    string             `bson:"user_name"`
	PhoneNumber string             `bson:"phone_number"`
}

type AlertRepository interface {
	Create(ctx context.Context, alert *Alert) error
	UpdateStatus(ctx context.Context, alertID string, status string) error
	FetchByID(ctx context.Context, alertID string) (*Alert, error)
	FetchByRadius(ctx context.Context, lat, lng, km float64) ([]Alert, error)
	GetNearbyAlerts(ctx context.Context, lat, lon, km float64) ([]*Alert, error)
}

type AlertUsecase interface {
	Handle(userID string, alert *Alert) error
	Raise(userID string, lat, lng, radius float64, ttl int, body string) (*Alert, error)
}
