package redis

import (
	"context"
	"fmt"

	"go-arch/internal/infra/configuration"

	"github.com/redis/go-redis/v9"
)

type Registry struct {
	clients map[string]*redis.Client
}

func NewRegistry(configs map[string]configuration.RedisConfig) (*Registry, error) {
	r := &Registry{
		clients: make(map[string]*redis.Client),
	}

	for name, cfg := range configs {
		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password: cfg.Password,
			DB:       cfg.DB,
		})

		if err := client.Ping(context.Background()).Err(); err != nil {
			r.Close()
			return nil, fmt.Errorf("pinging redis %q: %w", name, err)
		}

		r.clients[name] = client
	}

	return r, nil
}

func (r *Registry) Get(name string) (*redis.Client, error) {
	client, ok := r.clients[name]
	if !ok {
		return nil, fmt.Errorf("redis connection %q not found", name)
	}
	return client, nil
}

func (r *Registry) Close() error {
	var firstErr error
	for name, client := range r.clients {
		if err := client.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("closing redis %q: %w", name, err)
		}
	}
	return firstErr
}
