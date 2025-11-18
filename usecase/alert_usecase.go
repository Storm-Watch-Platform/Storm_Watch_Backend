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
func (uc *AlertUseCase) Handle(userID string, alert *domain.Alert) error {
	alert.UserID = userID

	job := worker.Job{
		Priority: 2, // ưu tiên cao
		Exec: func() {
			ctx, cancel := context.WithTimeout(context.Background(), uc.Timeout)
			defer cancel()

			// 1. Lưu alert vào DB
			if err := uc.Repo.Create(ctx, alert); err != nil {
				// có thể log lỗi
				return
			}

			// 2. Lấy danh sách alert trong bán kính 2 km (hardcode tạm hoặc dùng repo)
			// Lúc sau sẽ thay bằng query MongoDB
			users, err := uc.Repo.FetchByRadius(ctx, alert.Lat, alert.Lng, 2.0)
			if err != nil {
				return
			}

			// 3. Broadcast tới WSManager
			for _, u := range users {
				println(u.UserID)
				//uc.WSManager.Broadcast(u.UserID, alert.Body)
			}
		},
	}

	uc.Queue.Push(job)
	return nil
}
