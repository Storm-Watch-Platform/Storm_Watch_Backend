package route

import (
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/api/controller"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/bootstrap"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/repository"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/usecase"
	"github.com/gin-gonic/gin"
)

// NewCellRouter sets up routes for Cell operations
func NewCellRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	// Repository + Usecase
	cr := repository.NewCellRepository(db, domain.CollectionCell)
	cu := usecase.NewCellUsecase(cr, timeout)

	// Controller
	cc := &controller.CellController{
		CellUsecase: cu,
	}

	// --- Routes ---
	group.GET("/cell", cc.GetCellByLatLon)              // fetch 1 cell tại vị trí
	group.GET("/cells", cc.GetCellsByRadius)            // fetch nhiều cell theo radius
	group.POST("/cell", cc.UpdateCell)                  // gửi & chỉnh sửa 1 cell
	group.POST("/cells/radius", cc.UpdateCellsByRadius) // gửi & chỉnh sửa theo radius
}
