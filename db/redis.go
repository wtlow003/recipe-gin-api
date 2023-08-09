package databases

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis"
)

type RedisCache struct {
	Client *redis.Client
}

var (
	rPing = func(client *redis.Client) *redis.StatusCmd {
		return client.Ping()
	}
)

func ConnectToRedis(ctx context.Context, password, host, port string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address(host, port),
		Password: password,
		DB:       0,
	})
	// attempt to ping db
	if status := rPing(client); status.Val() != "PONG" {
		log.Fatalf("Failed to ping Redis cache, status = %s", status.Val())
	}
	log.Println("Connected to Redis!")
	return &RedisCache{
		Client: client,
	}, nil
}

func address(host, port string) string {
	const format = "%s:%s"
	return fmt.Sprintf(format, host, port)
}
