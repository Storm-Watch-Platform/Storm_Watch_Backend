package usecase

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/internal/ws"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/worker"
)

type ReportUseCase struct {
	Repo      domain.ReportRepository
	WSManager *ws.WSManager
	Queue     *worker.PriorityQueue
	Timeout   time.Duration
}

func NewReportUC(queue *worker.PriorityQueue, wsManager *ws.WSManager, repo domain.ReportRepository, timeout time.Duration) *ReportUseCase {
	return &ReportUseCase{
		Repo:      repo,
		WSManager: wsManager,
		Queue:     queue,
		Timeout:   timeout,
	}
}

// Handle nhận report từ FE
func (uc *ReportUseCase) Handle(userID string, report *domain.Report) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	report.UserID = id.Hex()

	job := worker.Job{
		Priority: 1, // priority trung bình
		Exec: func() {
			// 1. Lưu report vào DB
			_ = uc.Repo.Create(nil, report) // context nil tạm thời

			// 2. Hardcode danh sách user nhận broadcast
			users := []string{"userA", "userB"} // tạm thời

			// 3. Broadcast
			for _, u := range users {
				println("Broadcast report to", u)
				//uc.WSManager.Broadcast(u, report.Message)
			}
		},
	}

	uc.Queue.Push(job)
	return nil
}
