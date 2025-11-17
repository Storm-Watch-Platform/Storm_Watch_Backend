package main

import (
	"os"
	"time"

	route "github.com/Storm-Watch-Platform/Storm_Watch_Backend/api/route"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/bootstrap"
	"github.com/gin-gonic/gin"
)

func main() {

	// mở kết nối
	app := bootstrap.App()

	env := app.Env

	// lấy database từ cluster
	db := app.Mongo.Database(env.DBName)
	defer app.CloseDBConnection()

	timeout := time.Duration(env.ContextTimeout) * time.Second

	gin := gin.Default()

	route.Setup(env, timeout, db, gin)

	// gin.Run(env.ServerAddress)
	// Railway cung cấp port runtime qua env variable PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback khi chạy local
	}

	// Gin listen đúng port
	gin.Run(":" + port)
}
