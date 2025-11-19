package usecase

import (
	"context"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/internal/ai"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/internal/ws"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/worker"
)

type ReportUseCase struct {
	queue   *worker.PriorityQueue
	aiQueue *worker.AIQueue
	ws      *ws.WSManager
	repo    domain.ReportRepository
	timeout time.Duration
}

func NewReportUC(q *worker.PriorityQueue, aiq *worker.AIQueue, wsm *ws.WSManager, repo domain.ReportRepository, timeout time.Duration) *ReportUseCase {
	return &ReportUseCase{
		queue:   q,
		aiQueue: aiq,
		ws:      wsm,
		repo:    repo,
		timeout: timeout,
	}
}

func (uc *ReportUseCase) Handle(userID string, r *domain.Report) error {
	r.UserID = userID
	r.Timestamp = time.Now().Unix() // giống Location

	// Step 1: save + broadcast in PriorityQueue
	uc.queue.Push(worker.Job{
		Priority: 2, // Report > Location
		Exec: func() {
			ctx, cancel := context.WithTimeout(context.Background(), uc.timeout)
			defer cancel()

			// Save report
			_ = uc.repo.Create(ctx, r)
			// uc.ws.BroadcastReport(userID, r)
		},
	})

	// Step 2: AI analyze — run in separate AI queue
	uc.aiQueue.Push(func() {
		result, _ := ai.AnalyzeReport(r.Detail)

		// Lưu kết quả AI vào report
		ctx, cancel := context.WithTimeout(context.Background(), uc.timeout)
		defer cancel()

		r.Enrichment = &domain.ReportEnrichment{
			Category:    result.Category,
			Urgency:     result.Urgency,
			Summary:     result.Summary,
			Confidence:  result.Confidence,
			ExtractedAt: time.Now().Unix(),
		}

		// Update report trong DB với AI result
		if repoWithUpdate, ok := uc.repo.(interface {
			UpdateAI(ctx context.Context, reportID string, enrichment *domain.ReportEnrichment) error
		}); ok {
			_ = repoWithUpdate.UpdateAI(ctx, r.ID.Hex(), r.Enrichment)
		}

		// uc.ws.BroadcastAIResult(userID, result)
	})

	return nil
}

// Lấy report gần
func (uc *ReportUseCase) GetNearbyReports(ctx context.Context, lat, lon, km float64) ([]*domain.Report, error) {
	return uc.repo.GetNearbyReports(ctx, lat, lon, km)
}
