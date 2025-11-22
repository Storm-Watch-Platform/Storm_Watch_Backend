package route

import (
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/api/controller"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/bootstrap"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/internal/ws"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/repository"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/usecase"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/worker"
	"github.com/gin-gonic/gin"
)

func NewWSRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {

	// ================== //
	// 1. PRIORITY QUEUE (core realtime)
	// ================== //
	queue := worker.NewPriorityQueue()
	queue.Start(2) // 2 worker goroutines xử lý realtime job

	// ================== //
	// 2. ASYNC AI QUEUE (không priority)
	// ================== //
	aiQueue := worker.NewAIQueue()
	aiQueue.Start(1) // 1 worker chạy nhẹ thôi

	// ================== //
	// 3. WS MANAGER
	// ================== //
	wsManager := ws.NewWSManager()

	// ================== //
	// 4. REPOSITORIES
	// ================== //
	locRepo := repository.NewLocationRepo(db, domain.CollectionLocation)
	alertRepo := repository.NewAlertRepo(db, domain.CollectionAlert)
	reportRepo := repository.NewReportRepo(db, domain.CollectionReport)
	zoneRepo := repository.NewZoneRepository(db, domain.CollectionZone)

	// ================== //
	// 5. USE CASES
	// ================== //
	zoneUC := usecase.NewZoneUsecase(zoneRepo, timeout)
	locUC := usecase.NewLocationUC(queue, wsManager, locRepo, timeout)
	alertUC := usecase.NewAlertUC(queue, wsManager, alertRepo, locUC, timeout)
	reportUC := usecase.NewReportUC(queue, aiQueue, wsManager, reportRepo, zoneUC, timeout)

	// ================== //
	// 6. CONTROLLER
	// ================== //
	c := &controller.WSController{
		WSManager:  wsManager,
		LocationUC: locUC,
		AlertUC:    alertUC,
		ReportUC:   reportUC,
		ZoneUC:     zoneUC,
	}

	// ================== //
	// 7. ROUTE
	// ================== //
	group.GET("/ws", c.HandleWS)
}
