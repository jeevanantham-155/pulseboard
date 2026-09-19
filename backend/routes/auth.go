package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lords/live-polling/backend/controllers"
)

func RegisterAuthRoutes(router *gin.RouterGroup, controller *controllers.AuthController) {
	authRoutes := router.Group("/auth")
	authRoutes.POST("/signup", controller.Signup)
	authRoutes.POST("/login", controller.Login)
}
