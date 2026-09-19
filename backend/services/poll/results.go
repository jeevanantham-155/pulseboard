package poll

import (
	"github.com/lords/live-polling/backend/models"
	"time"
)

type PublicPoll struct {
	ID        string       `json:"id"`
	Question  string       `json:"question"`
	Options   []PollResult `json:"results"`
	Status    string       `json:"status"`
	ExpiresAt *time.Time   `json:"expiresAt,omitempty"`
}

type PollResult struct {
	OptionID   string  `json:"optionId"`
	Text       string  `json:"text"`
	Votes      int     `json:"votes"`
	Percentage float64 `json:"percentage"`
}

func ToPublicPoll(record models.Poll) PublicPoll {
	totalVotes := 0
	for _, option := range record.Options {
		totalVotes += option.VoteCount
	}
	results := make([]PollResult, 0, len(record.Options))
	for _, option := range record.Options {
		percentage := float64(0)
		if totalVotes > 0 {
			percentage = float64(option.VoteCount) / float64(totalVotes) * 100
		}
		results = append(results, PollResult{OptionID: option.ID, Text: option.Text, Votes: option.VoteCount, Percentage: percentage})
	}
	return PublicPoll{ID: record.ID.Hex(), Question: record.Question, Options: results, Status: record.Status, ExpiresAt: record.ExpiresAt}
}
