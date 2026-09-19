package poll

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/lords/live-polling/backend/models"
	"github.com/lords/live-polling/backend/repository"
	redisservice "github.com/lords/live-polling/backend/services/redis"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrPollInactive    = errors.New("poll is not active")
	ErrOptionInvalid   = errors.New("option is invalid")
	ErrVoterKeyInvalid = errors.New("voter key is invalid")
)

type VoteService struct {
	polls *repository.PollRepository
	votes *repository.VoteRepository
	redis *redisservice.Client
}

func NewVoteService(polls *repository.PollRepository, votes *repository.VoteRepository, redis *redisservice.Client) *VoteService {
	return &VoteService{polls: polls, votes: votes, redis: redis}
}

func (service *VoteService) Cast(ctx context.Context, pollID primitive.ObjectID, optionID string, voterKey string) (PublicPoll, error) {
	voterKey = strings.TrimSpace(voterKey)
	if len(voterKey) < 16 || len(voterKey) > 128 {
		return PublicPoll{}, ErrVoterKeyInvalid
	}
	pollRecord, err := service.polls.FindByID(ctx, pollID)
	if err != nil {
		return PublicPoll{}, err
	}
	if pollRecord.Status != models.PollStatusActive {
		return PublicPoll{}, ErrPollInactive
	}
	if pollRecord.ExpiresAt != nil && !pollRecord.ExpiresAt.After(time.Now().UTC()) {
		return PublicPoll{}, ErrPollInactive
	}
	optionExists := false
	for _, option := range pollRecord.Options {
		if option.ID == optionID {
			optionExists = true
			break
		}
	}
	if !optionExists {
		return PublicPoll{}, ErrOptionInvalid
	}

	vote := models.Vote{ID: primitive.NewObjectID(), PollID: pollID, OptionID: optionID, VoterKey: voterKey, CreatedAt: time.Now().UTC()}
	if err := service.votes.Create(ctx, vote); err != nil {
		return PublicPoll{}, err
	}
	if _, err := service.polls.IncrementOption(ctx, pollID, optionID); err != nil {
		_ = service.votes.Delete(ctx, vote.ID)
		return PublicPoll{}, err
	}
	updatedPoll, err := service.polls.FindByID(ctx, pollID)
	if err != nil {
		return PublicPoll{}, err
	}
	result := ToPublicPoll(updatedPoll)
	eventResults := make([]redisservice.Result, 0, len(result.Options))
	for _, option := range result.Options {
		eventResults = append(eventResults, redisservice.Result{OptionID: option.OptionID, Text: option.Text, Votes: option.Votes, Percentage: option.Percentage})
	}
	if err := service.redis.PublishResults(ctx, result.ID, eventResults); err != nil {
		return PublicPoll{}, err
	}
	return result, nil
}
