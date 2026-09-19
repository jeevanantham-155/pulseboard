package poll

import (
	"context"
	"time"

	"github.com/lords/live-polling/backend/services/redis"
)

type ExpiryWorker struct {
	polls *Service
	redis *redis.Client
	stop  chan struct{}
	done  chan struct{}
}

func NewExpiryWorker(polls *Service, redisClient *redis.Client) *ExpiryWorker {
	return &ExpiryWorker{polls: polls, redis: redisClient, stop: make(chan struct{}), done: make(chan struct{})}
}

func (worker *ExpiryWorker) Start() {
	go func() {
		defer close(worker.done)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case now := <-ticker.C:
				worker.closeExpired(now)
			case <-worker.stop:
				return
			}
		}
	}()
}

func (worker *ExpiryWorker) Stop() {
	close(worker.stop)
	<-worker.done
}

func (worker *ExpiryWorker) closeExpired(now time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	expired, err := worker.polls.CloseExpired(ctx, now.UTC())
	if err != nil {
		return
	}
	for _, record := range expired {
		_ = worker.redis.PublishClosed(ctx, record.ID.Hex())
	}
}
