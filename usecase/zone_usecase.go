package usecase

import (
	"context"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
)

type zoneUsecase struct {
	zoneRepository domain.ZoneRepository
	contextTimeout time.Duration
}

func NewZoneUsecase(repo domain.ZoneRepository, timeout time.Duration) domain.ZoneUsecase {
	return &zoneUsecase{
		zoneRepository: repo,
		contextTimeout: timeout,
	}
}

// Create tạo một zone mới
func (zu *zoneUsecase) Create(ctx context.Context, z *domain.Zone) error {
	ctx, cancel := context.WithTimeout(ctx, zu.contextTimeout)
	defer cancel()
	z.UpdatedAt = time.Now().UnixMilli()
	return zu.zoneRepository.Create(ctx, z)
}

// FetchAll lấy tất cả zone
func (zu *zoneUsecase) FetchAll(ctx context.Context) ([]domain.Zone, error) {
	ctx, cancel := context.WithTimeout(ctx, zu.contextTimeout)
	defer cancel()
	return zu.zoneRepository.FetchAll(ctx)
}

// FetchInBounds lấy các zone trong bounding box
func (zu *zoneUsecase) FetchInBounds(ctx context.Context, minLat, minLon, maxLat, maxLon float64) ([]domain.Zone, error) {
	ctx, cancel := context.WithTimeout(ctx, zu.contextTimeout)
	defer cancel()
	return zu.zoneRepository.FetchInBounds(ctx, minLat, minLon, maxLat, maxLon)
}

// FetchByLatLon lấy zone chứa một điểm cụ thể
func (zu *zoneUsecase) FetchAllByLatLon(ctx context.Context, lat, lon float64) ([]domain.Zone, error) {
	ctx2, cancel := context.WithTimeout(ctx, zu.contextTimeout)
	defer cancel()
	return zu.zoneRepository.FetchAllByLatLon(ctx2, lat, lon)
}

func (zu *zoneUsecase) AddRiskOrCreate(ctx context.Context, lat, lon, riskIncrement, defaultRadius float64) error {
	ctx2, cancel := context.WithTimeout(ctx, zu.contextTimeout)
	defer cancel()

	// 1️⃣ Lấy tất cả zone chứa điểm này
	zones, err := zu.zoneRepository.FetchAllByLatLon(ctx2, lat, lon)
	if err != nil {
		return err
	}

	if len(zones) == 0 {
		// 2️⃣ Nếu chưa có zone nào → tạo mới
		newZone := &domain.Zone{
			Center: domain.GeoPoint{
				Type:        "Point",
				Coordinates: [2]float64{lon, lat},
			},
			Radius:    defaultRadius,
			RiskScore: riskIncrement,
			Label:     "LOW",
			UpdatedAt: time.Now().UnixMilli(),
		}
		return zu.zoneRepository.Create(ctx2, newZone)
	}
	getLabel := func(riskscorein float64) string {
		switch {
		case riskscorein < 0.3:
			return "LOW"
		case riskscorein < 0.6:
			return "MEDIUM"
		default:
			return "HIGH"
		}
	}

	// 3️⃣ Nếu có zone → tăng risk
	for _, z := range zones {
		z.RiskScore += riskIncrement
		if z.RiskScore > 1 {
			z.RiskScore = 1
		}
		z.Label = getLabel(z.RiskScore)
		z.UpdatedAt = time.Now().UnixMilli()
		if err := zu.zoneRepository.Update(ctx2, &z); err != nil {
			return err
		}
	}

	return nil
}

func (zu *zoneUsecase) SetMaxRisk(ctx context.Context, lat, lon, newRisk float64) error {
	ctx2, cancel := context.WithTimeout(ctx, zu.contextTimeout)
	defer cancel()

	zones, err := zu.zoneRepository.FetchAllByLatLon(ctx2, lat, lon)
	if err != nil {
		return err
	}

	// Helper map risk -> label
	getLabel := func(r float64) string {
		switch {
		case r < 0.3:
			return "LOW"
		case r < 0.6:
			return "MEDIUM"
		default:
			return "HIGH"
		}
	}

	for _, z := range zones {
		if z.RiskScore < newRisk {
			z.RiskScore = newRisk
		} else {
			z.RiskScore += 0.1
			if z.RiskScore > 1.0 {
				z.RiskScore = 1.0
			}
		}

		z.Label = getLabel(z.RiskScore)
		z.UpdatedAt = time.Now().UnixMilli()

		if err := zu.zoneRepository.Update(ctx2, &z); err != nil {
			return err
		}
	}

	return nil
}
