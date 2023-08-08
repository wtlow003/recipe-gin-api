package db

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Database string
	Client   *mongo.Client
}

var (
	connect = func(ctx context.Context, client *mongo.Client) error {
		return client.Connect(ctx)
	}
	ping = func(ctx context.Context, client *mongo.Client) error {
		return client.Ping(ctx, nil)
	}
)

// Connects to a running MongoDB instance
func ConnectToMongoDB(ctx context.Context, user, password, host, database, port string) (*MongoDB, error) {
	client, err := mongo.NewClient(
		options.Client().ApplyURI(uri(user, password, host, database, port)),
	)
	if err != nil {
		log.Fatalf("Failed to create MongoDB client, err = %s", err)
	}
	// attempt to connect to db
	err = connect(ctx, client)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB server, err = %s", err)
	}
	err = ping(ctx, client)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB server, err = %s", err)
	}
	log.Println("Connected to MongoDB!")
	return &MongoDB{
		Database: database,
		Client:   client,
	}, nil
}

func uri(user, password, host, database, port string) string {
	const format = "mongodb://%s:%s@%s:%s/%s?authSource=admin"
	return fmt.Sprintf(format, user, password, host, port, database)
}
