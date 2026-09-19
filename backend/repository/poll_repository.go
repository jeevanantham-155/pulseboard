package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lords/live-polling/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrPollNotFound = errors.New("poll not found")

type PollRepository struct {
	collection *mongo.Collection
}

func NewPollRepository(database *mongo.Database) *PollRepository {
	return &PollRepository{collection: database.Collection("polls")}
}

func (repository *PollRepository) Create(ctx context.Context, poll models.Poll) error {
	_, err := repository.collection.InsertOne(ctx, poll)
	if err != nil {
		return fmt.Errorf("create poll: %w", err)
	}
	return nil
}

func (repository *PollRepository) FindByID(ctx context.Context, id primitive.ObjectID) (models.Poll, error) {
	var poll models.Poll
	if err := repository.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&poll); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.Poll{}, ErrPollNotFound
		}
		return models.Poll{}, fmt.Errorf("find poll: %w", err)
	}
	return poll, nil
}

func (repository *PollRepository) FindByCreator(ctx context.Context, creatorID primitive.ObjectID) ([]models.Poll, error) {
	cursor, err := repository.collection.Find(ctx, bson.M{"creatorId": creatorID}, optionsSortUpdated())
	if err != nil {
		return nil, fmt.Errorf("find creator polls: %w", err)
	}
	defer cursor.Close(ctx)

	var polls []models.Poll
	if err := cursor.All(ctx, &polls); err != nil {
		return nil, fmt.Errorf("decode creator polls: %w", err)
	}
	return polls, nil
}

func (repository *PollRepository) Delete(ctx context.Context, id primitive.ObjectID, creatorID primitive.ObjectID) error {
	result, err := repository.collection.DeleteOne(ctx, bson.M{"_id": id, "creatorId": creatorID})
	if err != nil {
		return fmt.Errorf("delete poll: %w", err)
	}
	if result.DeletedCount == 0 {
		return ErrPollNotFound
	}
	return nil
}

func (repository *PollRepository) CloseExpired(ctx context.Context, now time.Time) ([]models.Poll, error) {
	filter := bson.M{"status": models.PollStatusActive, "expiresAt": bson.M{"$lte": now}}
	cursor, err := repository.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("find expired polls: %w", err)
	}
	defer cursor.Close(ctx)
	var expired []models.Poll
	if err := cursor.All(ctx, &expired); err != nil {
		return nil, fmt.Errorf("decode expired polls: %w", err)
	}
	if len(expired) == 0 {
		return expired, nil
	}
	if _, err := repository.collection.UpdateMany(ctx, filter, bson.M{"$set": bson.M{"status": models.PollStatusClosed, "updatedAt": now}}); err != nil {
		return nil, fmt.Errorf("close expired polls: %w", err)
	}
	for index := range expired {
		expired[index].Status = models.PollStatusClosed
		expired[index].UpdatedAt = now
	}
	return expired, nil
}

func (repository *PollRepository) IncrementOption(ctx context.Context, pollID primitive.ObjectID, optionID string) (models.Poll, error) {
	result, err := repository.collection.UpdateOne(ctx,
		bson.M{"_id": pollID, "status": models.PollStatusActive, "options.id": optionID},
		bson.M{"$inc": bson.M{"options.$.voteCount": 1}},
	)
	if err != nil {
		return models.Poll{}, fmt.Errorf("increment poll option: %w", err)
	}
	if result.MatchedCount == 0 {
		return models.Poll{}, ErrPollNotFound
	}
	return repository.FindByID(ctx, pollID)
}
