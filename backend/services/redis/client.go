package redis

import (
	"context"
	"fmt"
	"time"

	redisclient "github.com/redis/go-redis/v9"
)

const defaultTimeout = 10 * time.Second

type Client struct {
	client *redisclient.Client
}

func Connect(ctx context.Context, redisURL string) (*Client, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("redis URL is required")
	}

	options, err := redisclient.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}

	redisClient := redisclient.NewClient(options)
	client := &Client{client: redisClient}
	if err := client.Ping(ctx); err != nil {
		_ = redisClient.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

func (client *Client) Ping(ctx context.Context) error {
	pingContext, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	return client.client.Ping(pingContext).Err()
}

func (client *Client) Close() error {
	return client.client.Close()
}
