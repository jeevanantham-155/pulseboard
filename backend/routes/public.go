package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lords/live-polling/backend/controllers"
)

func RegisterPublicRoutes(router *gin.RouterGroup, controller *controllers.PublicPollController) {
	publicRoutes := router.Group("/public/polls")
	publicRoutes.GET("/:id", controller.Get)
	publicRoutes.POST("/:id/vote", controller.Vote)
}
