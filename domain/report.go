package domain

import (
	"context"
)

const CollectionReport = "reports"

type Report struct {
	ID          string   `bson:"_id" json:"id"`
	UserID      string   `bson:"user_id" json:"user_id"`
	Type        string   `bson:"type" json:"type"`
	Detail      string   `bson:"detail" json:"detail"`
	Description string   `bson:"description" json:"description"`
	Image       string   `bson:"image" json:"image"`
	Location    GeoPoint `bson:"location" json:"location"`
	Timestamp   int64    `bson:"timestamp" json:"timestamp"`
	Status      string   `bson:"status" json:"status"`

	// AI OUTPUT
	Enrichment *ReportEnrichment `bson:"enrichment,omitempty" json:"enrichment,omitempty"`
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
	// FetchByGroupID(ctx context.Context, groupID string) ([]Report, error)
}

type ReportUser struct {
	UserID string
}

type ReportUsecase interface {
	Send(ctx context.Context, report *Report) error
	// FetchByGroupID(ctx context.Context, groupID string) ([]Report, error)
}
