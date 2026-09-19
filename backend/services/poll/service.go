package poll

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/lords/live-polling/backend/models"
	"github.com/lords/live-polling/backend/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrInvalidPoll = errors.New("invalid poll details")

type Service struct {
	polls *repository.PollRepository
}

func NewService(polls *repository.PollRepository) *Service {
	return &Service{polls: polls}
}

type Settings struct {
	ExpiresAt *time.Time
}

func (service *Service) Create(ctx context.Context, creatorID primitive.ObjectID, question string, optionTexts []string) (models.Poll, error) {
	return service.CreateWithSettings(ctx, creatorID, question, optionTexts, Settings{})
}

func (service *Service) CreateWithSettings(ctx context.Context, creatorID primitive.ObjectID, question string, optionTexts []string, settings Settings) (models.Poll, error) {
	question = strings.TrimSpace(question)
	if len(question) < 5 || len(question) > 500 || len(optionTexts) < 2 || len(optionTexts) > 10 {
		return models.Poll{}, ErrInvalidPoll
	}

	options := make([]models.PollOption, 0, len(optionTexts))
	seen := make(map[string]struct{}, len(optionTexts))
	for _, optionText := range optionTexts {
		optionText = strings.TrimSpace(optionText)
		key := strings.ToLower(optionText)
		if len(optionText) == 0 || len(optionText) > 200 {
			return models.Poll{}, ErrInvalidPoll
		}
		if _, exists := seen[key]; exists {
			return models.Poll{}, ErrInvalidPoll
		}
		seen[key] = struct{}{}
		options = append(options, models.PollOption{
			ID:        primitive.NewObjectID().Hex(),
			Text:      optionText,
			VoteCount: 0,
		})
	}

	now := time.Now().UTC()
	poll := models.Poll{
		ID:        primitive.NewObjectID(),
		CreatorID: creatorID,
		Question:  question,
		Options:   options,
		Status:    models.PollStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: settings.ExpiresAt,
	}
	if err := service.polls.Create(ctx, poll); err != nil {
		return models.Poll{}, err
	}
	return poll, nil
}

func (service *Service) CloseExpired(ctx context.Context, now time.Time) ([]models.Poll, error) {
	return service.polls.CloseExpired(ctx, now)
}

func (service *Service) ListByCreator(ctx context.Context, creatorID primitive.ObjectID) ([]models.Poll, error) {
	return service.polls.FindByCreator(ctx, creatorID)
}

func (service *Service) FindByID(ctx context.Context, id primitive.ObjectID) (models.Poll, error) {
	return service.polls.FindByID(ctx, id)
}

func (service *Service) Delete(ctx context.Context, id primitive.ObjectID, creatorID primitive.ObjectID) error {
	return service.polls.Delete(ctx, id, creatorID)
}
