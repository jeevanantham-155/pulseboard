package routes

import (
	"github.com/gin-gonic/gin"
	pollwebsocket "github.com/lords/live-polling/backend/websocket"
)

func RegisterWebSocketRoutes(router *gin.Engine, manager *pollwebsocket.Manager) {
	router.GET("/ws/polls/:pollId", manager.ServeHTTP)
}
