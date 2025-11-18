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
	// PRIORITY QUEUE
	// ================== //
	queue := worker.NewPriorityQueue()
	queue.Start(2) // 2 worker goroutine

	// ================== //
	// WS MANAGER
	// ================== //
	wsManager := ws.NewWSManager()
	// danh sách client, quản lý subscribe, broadcast các kiểu

	// ================== //
	// REPOSITORIES
	// ================== //
	locRepo := repository.NewLocationRepo(db, domain.CollectionLocation)
	alertRepo := repository.NewAlertRepo(db, domain.CollectionAlert)
	reportRepo := repository.NewReportRepo(db, domain.CollectionReport)
	// database cho loc, alert, report

	// ================== //
	// USE CASES
	// ================== //
	locUC := usecase.NewLocationUC(queue, wsManager, locRepo, timeout)
	alertUC := usecase.NewAlertUC(queue, wsManager, alertRepo, timeout)
	reportUC := usecase.NewReportUC(queue, wsManager, reportRepo, timeout)

	// ================== //
	// CONTROLLER
	// ================== //
	c := &controller.WSController{
		WSManager:  wsManager,
		LocationUC: locUC,
		AlertUC:    alertUC,
		ReportUC:   reportUC,
	}

	// ================== //
	// ROUTE
	// ================== //
	group.GET("/ws", c.HandleWS)
}
