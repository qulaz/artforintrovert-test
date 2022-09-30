package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/qulaz/artforintrovert-test/internal/config"
	"github.com/qulaz/artforintrovert-test/internal/entity"
	"github.com/qulaz/artforintrovert-test/internal/usecase/repo"
	"github.com/qulaz/artforintrovert-test/pkg/logging"
	"github.com/qulaz/artforintrovert-test/pkg/mongodb"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	mongo, err := mongodb.New(cfg.Database.DSN)
	if err != nil {
		panic(err)
	}

	mongoDatabase := mongo.Client().Database(cfg.Database.Name)

	productRepo := repo.NewMongoRepository(mongoDatabase, logging.NewDummyLogger())

	products := make([]*entity.Product, 50_000)

	for i := range products {
		products[i] = &entity.Product{
			Id:          primitive.NewObjectID(),
			Name:        gofakeit.Name(),
			Description: gofakeit.JobDescriptor(),
			Price:       int32(gofakeit.IntRange(1, math.MaxInt32)),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	if err := productRepo.CreateProducts(ctx, products); err != nil {
		panic(err)
	}

	fmt.Printf("Created %d products\n", len(products))

	defer func() {
		cancel()
		mongo.Close()
	}()
}
