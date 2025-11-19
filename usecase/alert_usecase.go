// ---------- usecase/alert_usecase.go ----------
package usecase

import (
	"context"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/internal/ws"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/worker"
)

type AlertUseCase struct {
	Repo      domain.AlertRepository
	WSManager *ws.WSManager
	Queue     *worker.PriorityQueue
	Timeout   time.Duration
}

func NewAlertUC(queue *worker.PriorityQueue, wsManager *ws.WSManager, repo domain.AlertRepository, timeout time.Duration) *AlertUseCase {
	return &AlertUseCase{
		Repo:      repo,
		WSManager: wsManager,
		Queue:     queue,
		Timeout:   timeout,
	}
}

// Handle nhận alert từ FE
func (uc *AlertUseCase) Handle(c *ws.Client, alert *domain.Alert) error {
	alert.UserID = c.UserID
	// ... tạo job
	uc.Queue.Push(worker.Job{
		Priority: 20,
		Exec: func() {
			ctx, cancel := context.WithTimeout(context.Background(), uc.Timeout)
			defer cancel()

			if err := uc.Repo.Create(ctx, alert); err != nil {
				println("Failed to save alert:", err.Error())
				return
			}

			// gửi về tab hiện tại
			response := map[string]interface{}{
				"status":     "ok",
				"alertId":    alert.ID.Hex(),
				"expires_at": alert.ExpiresAt.Unix(),
			}
			uc.WSManager.SendToClient(c, "alert_response", response)
		},
	})
	return nil
}

func (uc *AlertUseCase) Resolve(client *ws.Client, alertID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), uc.Timeout)
	defer cancel()

	// 1. Update status trong DB
	err := uc.Repo.UpdateStatus(ctx, alertID, "RESOLVED")
	if err != nil {
		return err
	}

	// 2. Chuẩn bị response gửi về FE
	response := map[string]interface{}{
		"status":    "ok",
		"alertId":   alertID,
		"newStatus": "RESOLVED",
	}

	// 3. Gửi về WS của chính client
	uc.WSManager.SendToClient(client, "alert_resolved", response)

	return nil
}
