package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lords/live-polling/backend/controllers"
	"github.com/lords/live-polling/backend/middleware"
	"github.com/lords/live-polling/backend/services/auth"
)

func RegisterPollRoutes(router *gin.RouterGroup, controller *controllers.PollController, authService *auth.Service) {
	pollRoutes := router.Group("/polls")
	pollRoutes.Use(middleware.RequireAuth(authService))
	pollRoutes.POST("", controller.Create)
	pollRoutes.GET("", controller.List)
	pollRoutes.GET("/:id/results/pdf", controller.ResultsPDF)
	pollRoutes.GET("/:id", controller.Get)
	pollRoutes.DELETE("/:id", controller.Delete)
}
