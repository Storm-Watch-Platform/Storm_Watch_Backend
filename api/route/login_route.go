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

func NewLoginRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	ur := repository.NewUserRepository(db, domain.CollectionUser) // tạo user repository để làm việc với collection user
	lc := &controller.LoginController{                            // tạo controller
		LoginUsecase: usecase.NewLoginUsecase(ur, timeout),
		Env:          env,
	}
	group.POST("/login", lc.Login) // tạo
}
