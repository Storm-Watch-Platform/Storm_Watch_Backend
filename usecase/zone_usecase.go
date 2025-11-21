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
