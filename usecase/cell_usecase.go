package usecase

import (
	"context"
	"math"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
)

type CellUsecaseImpl struct {
	repo    domain.CellRepository
	timeout time.Duration
}

func NewCellUsecase(repo domain.CellRepository, timeout time.Duration) domain.CellUsecase {
	return &CellUsecaseImpl{
		repo:    repo,
		timeout: timeout,
	}
}

// snapGrid chuẩn hóa lat/lon thành cell ~100x100m
func snapGrid(lat, lon float64) (float64, float64) {
	cellSize := 0.001 // ~111m
	return float64(int(lat/cellSize)) * cellSize, float64(int(lon/cellSize)) * cellSize
}

// distanceMeters tính khoảng cách Haversine
func distanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // bán kính Trái Đất (m)
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

// ----------------------
// Usecase methods
// ----------------------

func (uc *CellUsecaseImpl) FetchByLatLon(ctx context.Context, lat, lon float64) (*domain.Cell, error) {
	latSnap, lonSnap := snapGrid(lat, lon)
	cell, err := uc.repo.GetByLatLon(ctx, latSnap, lonSnap)
	if err != nil {
		return nil, err
	}
	if cell == nil {
		cell = &domain.Cell{
			Center:    [2]float64{lonSnap, latSnap},
			RiskScore: 0,
			Label:     "SAFE",
			UpdatedAt: time.Now(),
		}
	}
	return cell, nil
}

func (uc *CellUsecaseImpl) GetCellByLatLon(ctx context.Context, lat, lon float64) (*domain.Cell, error) {
	latSnap, lonSnap := snapGrid(lat, lon)
	return uc.repo.GetByLatLon(ctx, latSnap, lonSnap)
}

func (uc *CellUsecaseImpl) GetCellsByRadius(ctx context.Context, lat, lon, radius float64) ([]domain.Cell, error) {
	return uc.repo.GetByRadius(ctx, lat, lon, radius)
}

func (uc *CellUsecaseImpl) UpdateCell(ctx context.Context, lat, lon, risk float64, label string) (*domain.Cell, error) {
	latSnap, lonSnap := snapGrid(lat, lon)
	cell := &domain.Cell{
		Center:    [2]float64{lonSnap, latSnap},
		RiskScore: risk,
		Label:     label,
		UpdatedAt: time.Now(),
	}
	return cell, uc.repo.Upsert(ctx, cell)
}

func (uc *CellUsecaseImpl) Upsert(ctx context.Context, cell *domain.Cell) error {
	return uc.repo.Upsert(ctx, cell)
}

func (uc *CellUsecaseImpl) UpdateCellsByRadius(ctx context.Context, lat, lon, radius, riskInc float64, label string) (updated int, inserted int, err error) {
	const cellSize = 0.001 // ~111m
	// 1️⃣ Lấy các cell hiện có
	cells, err := uc.repo.GetByRadius(ctx, lat, lon, radius)
	if err != nil {
		return 0, 0, err
	}
	existing := make(map[[2]float64]*domain.Cell)
	for i := range cells {
		c := &cells[i]
		existing[c.Center] = c
	}

	// 2️⃣ Tính số bước lặp theo cellSize
	steps := int(math.Ceil(radius / cellSize))

	for dx := -steps; dx <= steps; dx++ {
		for dy := -steps; dy <= steps; dy++ {
			cLat := lat + float64(dx)*cellSize
			cLon := lon + float64(dy)*cellSize
			if distanceMeters(lat, lon, cLat, cLon) > radius {
				continue
			}

			latNorm, lonNorm := snapGrid(cLat, cLon)
			center := [2]float64{lonNorm, latNorm}

			if cell, ok := existing[center]; ok {
				cell.RiskScore += riskInc
				if label != "" {
					cell.Label = label
				}
				_ = uc.repo.Upsert(ctx, cell)
				updated++
			} else {
				newCell := &domain.Cell{
					Center:    center,
					RiskScore: riskInc,
					Label:     label,
					UpdatedAt: time.Now(),
				}
				_ = uc.repo.Upsert(ctx, newCell)
				inserted++
			}
		}
	}
	return updated, inserted, nil
}
