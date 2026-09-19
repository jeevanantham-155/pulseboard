package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lords/live-polling/backend/repository"
	"github.com/lords/live-polling/backend/services/auth"
)

type AuthController struct {
	service *auth.Service
}

type signupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthController(service *auth.Service) *AuthController {
	return &AuthController{service: service}
}

func (controller *AuthController) Signup(context *gin.Context) {
	var request signupRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		writeError(context, http.StatusBadRequest, "request body must be valid JSON")
		return
	}

	user, token, err := controller.service.Signup(context.Request.Context(), request.Name, request.Email, request.Password)
	if err != nil {
		controller.writeServiceError(context, err)
		return
	}
	context.JSON(http.StatusCreated, gin.H{"user": user, "token": token})
}

func (controller *AuthController) Login(context *gin.Context) {
	var request loginRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		writeError(context, http.StatusBadRequest, "request body must be valid JSON")
		return
	}

	user, token, err := controller.service.Login(context.Request.Context(), request.Email, request.Password)
	if err != nil {
		controller.writeServiceError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}

func (controller *AuthController) writeServiceError(context *gin.Context, err error) {
	switch {
	case errors.Is(err, auth.ErrEmailExists):
		writeError(context, http.StatusConflict, "email is already registered")
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(context, http.StatusUnauthorized, "invalid email or password")
	case errors.Is(err, auth.ErrInvalidSignup):
		writeError(context, http.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrUserNotFound):
		writeError(context, http.StatusUnauthorized, "invalid email or password")
	default:
		writeError(context, http.StatusInternalServerError, "unable to process request")
	}
}

func writeError(context *gin.Context, status int, message string) {
	context.JSON(status, gin.H{"success": false, "message": message})
}
