package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionAlert = "alerts"

type Alert struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	UserID string             `bson:"userID"`
	Lat    float64            `bson:"lat"`
	Lng    float64            `bson:"lng"`
	Body   string             `bson:"body"`
}

type AlertRepository interface {
	Create(ctx context.Context, alert *Alert) error
	FetchByRadius(ctx context.Context, lat, lng, km float64) ([]Alert, error)
}

type AlertUsecase interface {
	Handle(userID string, alert *Alert) error
}
