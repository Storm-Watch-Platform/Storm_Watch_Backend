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
	"github.com/gin-gonic/gin"
)

func NewNearbyRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	// -----------------------
	// 1️⃣ Repositories
	// -----------------------
	alertRepo := repository.NewAlertRepo(db, domain.CollectionAlert)
	reportRepo := repository.NewReportRepo(db, domain.CollectionReport)
	locRepo := repository.NewLocationRepo(db, domain.CollectionLocation)
	zoneRepo := repository.NewZoneRepository(db, domain.CollectionZone)

	// -----------------------
	// 2️⃣ UseCases
	// -----------------------
	locUC := usecase.NewLocationUC(nil, nil, locRepo, timeout)
	alertUC := usecase.NewAlertUC(nil, nil, alertRepo, locUC, timeout)
	zoneUC := usecase.NewZoneUsecase(zoneRepo, timeout)
	reportUC := usecase.NewReportUC(nil, nil, nil, reportRepo, zoneUC, timeout)

	// -----------------------
	// 3️⃣ WS Manager (optional, nếu cần broadcast realtime)
	// -----------------------
	wsManager := ws.NewWSManager()

	// -----------------------
	// 4️⃣ NearbyController
	// -----------------------
	nc := controller.NewNearbyController(alertUC, reportUC, wsManager)

	// -----------------------
	// 5️⃣ Routes
	// -----------------------
	group.GET("/nearby/sos", nc.NearbySOS)
	group.GET("/nearby/report", nc.NearbyReport)
}
