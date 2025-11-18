package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/internal/ws"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/worker"
)

var allowedStatus = map[string]bool{
	"SAFE":    true,
	"CAUTION": true,
	"DANGER":  true,
	"UNKNOWN": true,
}

type LocationUseCase struct {
	queue   *worker.PriorityQueue
	ws      *ws.WSManager
	repo    domain.LocationRepository
	timeout time.Duration
}

func NewLocationUC(q *worker.PriorityQueue, wsm *ws.WSManager, repo domain.LocationRepository, timeout time.Duration) *LocationUseCase {
	return &LocationUseCase{
		queue:   q,
		ws:      wsm,
		repo:    repo,
		timeout: timeout,
	}
}

func (uc *LocationUseCase) Handle(userID string, loc *domain.Location) error {
	println("VAO HAM HANDLE CUA LOCATION USECASE")
	if !allowedStatus[loc.Status] {
		return errors.New("invalid status")
	}

	loc.ID = userID // _id = userID
	loc.UpdatedAt = time.Now().Unix()

	println("PUSH VÔ QUEUE")
	// push job vào worker queue
	uc.queue.Push(worker.Job{
		Priority: 1, // hoặc 0 tùy bạn muốn
		Exec: func() {
			ctx, cancel := context.WithTimeout(context.Background(), uc.timeout)
			defer cancel()

			// Lưu location
			_ = uc.repo.Upsert(ctx, loc)

			// Broadcast location
			uc.ws.BroadcastLocation(userID, loc)
		},
	})

	return nil
}
