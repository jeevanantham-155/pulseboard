package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultTimeout = 10 * time.Second

type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

func Connect(ctx context.Context, uri string, databaseName string) (*Client, error) {
	if uri == "" || databaseName == "" {
		return nil, fmt.Errorf("mongodb URI and database name are required")
	}

	connectContext, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	databaseClient, err := mongo.Connect(connectContext, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect to mongodb: %w", err)
	}

	if err := databaseClient.Ping(connectContext, nil); err != nil {
		_ = databaseClient.Disconnect(context.Background())
		return nil, fmt.Errorf("ping mongodb: %w", err)
	}

	return &Client{
		client:   databaseClient,
		database: databaseClient.Database(databaseName),
	}, nil
}

func (client *Client) Ping(ctx context.Context) error {
	pingContext, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	return client.client.Ping(pingContext, nil)
}

func (client *Client) Database() *mongo.Database {
	return client.database
}

func (client *Client) Disconnect(ctx context.Context) error {
	disconnectContext, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	return client.client.Disconnect(disconnectContext)
}
