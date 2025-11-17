package route

import (
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/api/controller"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/bootstrap"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/internal/ws"
	"github.com/gin-gonic/gin"
)

func NewWSRouter(env *bootstrap.Env, timeout time.Duration, group *gin.RouterGroup) {

	// controller: tập hợp các handler cho WS
	c := &controller.WSController{
		// manager để quản lý các kết nối WS (state, mem, cache, ...)
		WSManager: ws.NewWSManager(),
	}

	// STOMP entrypoint
	// handler: xử lý request STOMP frames
	group.GET("/ws", c.HandleWS)

	// backup API if you need HTTP
	// group.POST("/location", c.ReceiveLocation)
}
