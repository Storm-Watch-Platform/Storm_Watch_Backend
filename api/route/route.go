package route

import (
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/api/middleware"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/bootstrap"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"github.com/gin-gonic/gin"
)

func Setup(env *bootstrap.Env, timeout time.Duration, db mongo.Database, gin *gin.Engine) {
	publicRouter := gin.Group("")
	// All Public APIs
	NewSignupRouter(env, timeout, db, publicRouter)
	NewLoginRouter(env, timeout, db, publicRouter)
	NewRefreshTokenRouter(env, timeout, db, publicRouter)

	protectedRouter := gin.Group("")
	// Middleware to verify AccessToken
	// use để gắn middleware: chạy trước khi request tới route
	protectedRouter.Use(middleware.JwtAuthMiddleware(env.AccessTokenSecret))
	// All Private APIs -> nhóm API yêu cầu xác thực bằng AccessToken
	NewProfileRouter(env, timeout, db, protectedRouter)
	NewTaskRouter(env, timeout, db, protectedRouter)
}
