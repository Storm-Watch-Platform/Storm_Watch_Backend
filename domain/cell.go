package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	CollectionCell = "cells"
)

// Cell represents a grid cell for danger zones
type Cell struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Center    [2]float64         `bson:"center" json:"center"` // [lon, lat] để index 2dsphere
	RiskScore float64            `bson:"riskScore" json:"riskScore"`
	Label     string             `bson:"label" json:"label"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// ----------------------
// Repository interface
// ----------------------
type CellRepository interface {
	GetByLatLon(ctx context.Context, lat, lon float64) (*Cell, error)
	GetByRadius(ctx context.Context, lat, lon, radius float64) ([]Cell, error)
	Upsert(ctx context.Context, cell *Cell) error
	UpsertMany(ctx context.Context, cells []*Cell) error
}

// ----------------------
// Usecase interface
// ----------------------
type CellUsecase interface {
	FetchByLatLon(ctx context.Context, lat, lon float64) (*Cell, error)
	GetCellByLatLon(ctx context.Context, lat, lon float64) (*Cell, error)
	GetCellsByRadius(ctx context.Context, lat, lon, radius float64) ([]Cell, error)
	UpdateCell(ctx context.Context, lat, lon, risk float64, label string) (*Cell, error)
	UpdateCellsByRadius(ctx context.Context, lat, lon, radius, riskInc float64, label string) (updated int, inserted int, err error)
	Upsert(ctx context.Context, cell *Cell) error
}
