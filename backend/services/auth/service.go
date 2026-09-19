package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lords/live-polling/backend/models"
	"github.com/lords/live-polling/backend/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidSignup      = errors.New("invalid signup details")
)

type Service struct {
	users     *repository.UserRepository
	jwtSecret []byte
}

func NewService(users *repository.UserRepository, jwtSecret string) *Service {
	return &Service{users: users, jwtSecret: []byte(jwtSecret)}
}

func (service *Service) Signup(ctx context.Context, name string, email string, password string) (models.User, string, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	if err := validateSignup(name, email, password); err != nil {
		return models.User{}, "", err
	}

	_, err := service.users.FindByEmail(ctx, email)
	if err == nil {
		return models.User{}, "", ErrEmailExists
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return models.User{}, "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, "", fmt.Errorf("hash password: %w", err)
	}

	user := models.User{
		ID:           primitive.NewObjectID(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}
	if err := service.users.Create(ctx, user); err != nil {
		if strings.Contains(err.Error(), "E11000") {
			return models.User{}, "", ErrEmailExists
		}
		return models.User{}, "", err
	}

	token, err := service.createToken(user)
	return user, token, err
}

func (service *Service) Login(ctx context.Context, email string, password string) (models.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return models.User{}, "", ErrInvalidCredentials
	}

	user, err := service.users.FindByEmail(ctx, email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return models.User{}, "", ErrInvalidCredentials
	}

	token, err := service.createToken(user)
	return user, token, err
}

func (service *Service) ParseToken(tokenString string) (primitive.ObjectID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected token signing method")
		}
		return service.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return primitive.NilObjectID, ErrInvalidCredentials
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return primitive.NilObjectID, ErrInvalidCredentials
	}
	subject, err := claims.GetSubject()
	if err != nil {
		return primitive.NilObjectID, ErrInvalidCredentials
	}
	userID, err := primitive.ObjectIDFromHex(subject)
	if err != nil {
		return primitive.NilObjectID, ErrInvalidCredentials
	}
	return userID, nil
}

func (service *Service) createToken(user models.User) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   user.ID.Hex(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(service.jwtSecret)
}

func validateSignup(name string, email string, password string) error {
	if len(name) < 2 || len(name) > 80 {
		return fmt.Errorf("%w: name must be between 2 and 80 characters", ErrInvalidSignup)
	}
	if !strings.Contains(email, "@") || len(email) > 254 {
		return fmt.Errorf("%w: a valid email is required", ErrInvalidSignup)
	}
	if len(password) < 8 || len(password) > 72 {
		return fmt.Errorf("%w: password must be between 8 and 72 characters", ErrInvalidSignup)
	}
	return nil
}
