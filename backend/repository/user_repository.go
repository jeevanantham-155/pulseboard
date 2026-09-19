package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/lords/live-polling/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(database *mongo.Database) *UserRepository {
	return &UserRepository{collection: database.Collection("users")}
}

func (repository *UserRepository) Create(ctx context.Context, user models.User) error {
	_, err := repository.collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (repository *UserRepository) FindByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	if err := repository.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (repository *UserRepository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: optionsUnique(),
	})
	if err != nil {
		return fmt.Errorf("create user indexes: %w", err)
	}
	return nil
}
