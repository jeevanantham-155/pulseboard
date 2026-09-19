package redis

import (
	"context"
	"encoding/json"
	"fmt"

	redisclient "github.com/redis/go-redis/v9"
)

type Result struct {
	OptionID   string  `json:"optionId"`
	Text       string  `json:"text"`
	Votes      int     `json:"votes"`
	Percentage float64 `json:"percentage"`
}

type ResultsEvent struct {
	Type    string   `json:"type"`
	PollID  string   `json:"pollId"`
	Results []Result `json:"results"`
}

type PollClosedEvent struct {
	Type   string `json:"type"`
	PollID string `json:"pollId"`
	Status string `json:"status"`
}

func ResultsChannel(pollID string) string {
	return "poll:" + pollID + ":results"
}

func (client *Client) PublishResults(ctx context.Context, pollID string, results []Result) error {
	event, err := json.Marshal(ResultsEvent{Type: "poll_results_updated", PollID: pollID, Results: results})
	if err != nil {
		return fmt.Errorf("encode results event: %w", err)
	}
	return client.client.Publish(ctx, ResultsChannel(pollID), event).Err()
}

func (client *Client) PublishClosed(ctx context.Context, pollID string) error {
	event, err := json.Marshal(PollClosedEvent{Type: "poll_closed", PollID: pollID, Status: "closed"})
	if err != nil {
		return fmt.Errorf("encode closed event: %w", err)
	}
	return client.client.Publish(ctx, ResultsChannel(pollID), event).Err()
}

func (client *Client) SubscribeResults(ctx context.Context, pollID string) *redisclient.PubSub {
	return client.client.Subscribe(ctx, ResultsChannel(pollID))
}
