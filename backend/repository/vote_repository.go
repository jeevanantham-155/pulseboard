package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/lords/live-polling/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrDuplicateVote = errors.New("duplicate vote")

type VoteRepository struct {
	collection *mongo.Collection
}

func NewVoteRepository(database *mongo.Database) *VoteRepository {
	return &VoteRepository{collection: database.Collection("votes")}
}

func (repository *VoteRepository) Create(ctx context.Context, vote models.Vote) error {
	_, err := repository.collection.InsertOne(ctx, vote)
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicateVote
	}
	if err != nil {
		return fmt.Errorf("create vote: %w", err)
	}
	return nil
}

func (repository *VoteRepository) Delete(ctx context.Context, voteID interface{}) error {
	_, err := repository.collection.DeleteOne(ctx, bson.M{"_id": voteID})
	return err
}

func (repository *VoteRepository) EnsureIndexes(ctx context.Context) error {
	unique := true
	_, err := repository.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "pollId", Value: 1}, {Key: "voterKey", Value: 1}},
		Options: &options.IndexOptions{Unique: &unique},
	})
	if err != nil {
		return fmt.Errorf("create vote indexes: %w", err)
	}
	return nil
}
