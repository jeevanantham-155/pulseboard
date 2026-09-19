package poll

import (
	"testing"

	"github.com/lords/live-polling/backend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestToPublicPollCalculatesTextOptionPercentages(t *testing.T) {
	record := models.Poll{
		ID:       primitive.NewObjectID(),
		Question: "Choose one",
		Status:   models.PollStatusActive,
		Options: []models.PollOption{
			{ID: "one", Text: "One", VoteCount: 3},
			{ID: "two", Text: "Two", VoteCount: 1},
		},
	}
	result := ToPublicPoll(record)
	if result.Options[0].Text != "One" || result.Options[0].Percentage != 75 || result.Options[1].Percentage != 25 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestInvalidTextPollDoesNotWrite(t *testing.T) {
	service := NewService(nil)
	if _, err := service.Create(nil, primitive.NewObjectID(), "", []string{"only"}); err != ErrInvalidPoll {
		t.Fatalf("expected ErrInvalidPoll, got %v", err)
	}
}
