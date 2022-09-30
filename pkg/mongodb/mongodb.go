package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

const connectionTimeout = time.Second * 5

type MongoDB struct {
	client *mongo.Client
}

func New(dsn string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	opts := options.Client()
	opts.Monitor = otelmongo.NewMonitor()
	opts.ApplyURI(dsn)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("can't connect to mongo: %w", err)
	}

	for connectionTry := 0; connectionTry < 3; connectionTry++ {
		err = pingMongoDB(client)
		if err == nil {
			return &MongoDB{
				client: client,
			}, nil
		}
	}

	return nil, fmt.Errorf("error while ping mongodb client: %w", err)
}

func (m *MongoDB) Client() *mongo.Client {
	return m.client
}

func (m *MongoDB) Close() error {
	if m == nil || m.client == nil {
		return errors.New("mongodb client is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	return m.client.Disconnect(ctx)
}

func pingMongoDB(client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	return client.Ping(ctx, nil)
}
