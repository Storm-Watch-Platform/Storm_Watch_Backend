// package route

// import (
// 	"time"

// 	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/api/middleware"
// 	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/bootstrap"
// 	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
// 	"github.com/gin-gonic/gin"
// )

// func Setup(env *bootstrap.Env, timeout time.Duration, db mongo.Database, gin *gin.Engine) {
// 	publicRouter := gin.Group("")
// 	// All Public APIs
// 	NewSignupRouter(env, timeout, db, publicRouter)
// 	NewLoginRouter(env, timeout, db, publicRouter)
// 	NewRefreshTokenRouter(env, timeout, db, publicRouter)

// 	protectedRouter := gin.Group("")
// 	// Middleware to verify AccessToken
// 	protectedRouter.Use(middleware.JwtAuthMiddleware(env.AccessTokenSecret))
// 	// All Private APIs
// 	NewProfileRouter(env, timeout, db, protectedRouter)
// 	NewTaskRouter(env, timeout, db, protectedRouter)
// }

package route

import (
	"context"
	"net/http"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/api/middleware"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/bootstrap"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func Setup(env *bootstrap.Env, timeout time.Duration, db mongo.Database, router *gin.Engine) {
	publicRouter := router.Group("")
	// All Public APIs
	NewSignupRouter(env, timeout, db, publicRouter)
	NewLoginRouter(env, timeout, db, publicRouter)
	NewRefreshTokenRouter(env, timeout, db, publicRouter)

	protectedRouter := router.Group("")
	protectedRouter.Use(middleware.JwtAuthMiddleware(env.AccessTokenSecret))
	NewProfileRouter(env, timeout, db, protectedRouter)
	NewTaskRouter(env, timeout, db, protectedRouter)

	// ====== TEST ROUTES ======
	router.POST("/test-insert", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		collection := db.Collection("test_collection")

		doc := bson.M{
			"name":  "StormSafe",
			"role":  "admin",
			"time":  time.Now(),
			"notes": "Inserted via Go backend ✅",
		}

		_, err := collection.InsertOne(ctx, doc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Inserted test document!"})
	})

	router.GET("/test-find", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		collection := db.Collection("test_collection")

		var result bson.M
		err := collection.FindOne(ctx, bson.M{"name": "StormSafe"}).Decode(&result)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)

	})
}
