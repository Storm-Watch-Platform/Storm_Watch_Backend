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

func NewZoneRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	zr := repository.NewZoneRepository(db, domain.CollectionZone)

	zc := &controller.ZoneController{
		ZoneUsecase: usecase.NewZoneUsecase(zr, timeout),
	}

	group.POST("/zones", zc.Create)
	group.GET("/zones", zc.FetchInBounds)
	group.GET("/zones/by-location", zc.FetchByLatLon)
	group.GET("/zones/all", zc.FetchAll)
}
