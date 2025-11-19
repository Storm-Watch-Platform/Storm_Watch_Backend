package route

import (
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/api/controller"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/bootstrap"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/repository"
	"github.com/gin-gonic/gin"
)

// Chỉ dùng để mock dữ liệu report
func NewReportMockRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	// -----------------------
	// 1️⃣ Repository
	// -----------------------
	reportRepo := repository.NewReportRepo(db, "reports")

	// -----------------------
	// 2️⃣ Controller (dùng Repo trực tiếp)
	// -----------------------
	rc := &controller.ReportController{
		ReportRepo: reportRepo,
		Timeout:    timeout,
	}

	// -----------------------
	// 3️⃣ Route
	// -----------------------
	group.POST("/report/mock", rc.MockReport)
}

func NewAlertMockRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	alertRepo := repository.NewAlertRepo(db, "alerts")
	ac := controller.NewAlertController(alertRepo)
	group.POST("/alert/mock", ac.MockAlert)
}
