package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionReport = "reports"

type Report struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	Type        string             `bson:"type" json:"type"`
	Detail      string             `bson:"detail" json:"detail"`
	Description string             `bson:"description" json:"description"`
	Image       string             `bson:"image,omitempty" json:"image,omitempty"` // Base64
	Location    GeoPoint           `bson:"location" json:"location"`

	Timestamp   int64             `bson:"timestamp" json:"timestamp"`
	Status      string            `bson:"status" json:"status"`
	PhoneNumber string            `bson:"phone_number" json:"phone_number"`
	UserName    string            `bson:"user_name" json:"user_name"`
	Enrichment  *ReportEnrichment `bson:"enrichment,omitempty" json:"enrichment,omitempty"`
}
type ReportEnrichment struct {
	Category    string `bson:"category" json:"category"`         // “flood”, “fire”, “accident”...
	Urgency     string `bson:"urgency" json:"urgency"`           // “LOW”, “MEDIUM”, “HIGH”
	Summary     string `bson:"summary" json:"summary"`           // AI Summary
	Confidence  int    `bson:"confidence" json:"confidence"`     // 0–100
	ExtractedAt int64  `bson:"extracted_at" json:"extracted_at"` // timestamp
}

type ReportRepository interface {
	Create(ctx context.Context, report *Report) error
	GetNearbyReports(ctx context.Context, lat, lon, km float64) ([]*Report, error)
	// FetchByGroupID(ctx context.Context, groupID string) ([]Report, error)
}

type ReportUser struct {
	UserID string
}

type ReportUsecase interface {
	Send(ctx context.Context, report *Report) error
	GetNearbyReports(ctx context.Context, lat, lon, km float64) ([]*Report, error)
	// FetchByGroupID(ctx context.Context, groupID string) ([]Report, error)
}
