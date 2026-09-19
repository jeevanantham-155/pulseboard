package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lords/live-polling/backend/services/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const userIDKey = "userID"

func RequireAuth(service *auth.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		authorization := context.GetHeader("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "authentication required"})
			return
		}

		userID, err := service.ParseToken(strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer ")))
		if err != nil {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "invalid or expired token"})
			return
		}

		context.Set(userIDKey, userID)
		context.Next()
	}
}

func UserID(context *gin.Context) (primitive.ObjectID, bool) {
	value, exists := context.Get(userIDKey)
	if !exists {
		return primitive.NilObjectID, false
	}
	userID, ok := value.(primitive.ObjectID)
	return userID, ok
}
