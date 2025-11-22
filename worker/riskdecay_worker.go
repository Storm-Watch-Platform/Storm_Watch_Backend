package worker

import (
	"context"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
)

type RiskDecayWorker struct {
	zoneUC   domain.ZoneUsecase
	interval time.Duration
}

func NewRiskDecayWorker(zoneUC domain.ZoneUsecase, interval time.Duration) *RiskDecayWorker {
	return &RiskDecayWorker{
		zoneUC:   zoneUC,
		interval: interval,
	}
}

func (w *RiskDecayWorker) Start() {
	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for range ticker.C {
			w.DecayAllZones()
		}
	}()
}

// DecayAllZones chạy giảm risk cho tất cả zone
func (w *RiskDecayWorker) DecayAllZones() {
	ctx := context.Background()
	zones, err := w.zoneUC.FetchAll(ctx)
	if err != nil {
		return
	}

	for _, z := range zones {
		newRisk := z.RiskScore - 0.1
		if newRisk <= 0 {
			// Xóa zone nếu risk <= 0
			_ = w.zoneUC.Delete(ctx, z.ID)
			continue
		}

		if newRisk > 1.0 {
			newRisk = 1.0
		}

		// Cập nhật risk + label
		z.RiskScore = newRisk
		z.Label = getLabel(newRisk)
		_ = w.zoneUC.Update(ctx, &z)
	}
}

// Helper label
func getLabel(r float64) string {
	switch {
	case r < 0.3:
		return "LOW"
	case r < 0.6:
		return "MEDIUM"
	default:
		return "HIGH"
	}
}
