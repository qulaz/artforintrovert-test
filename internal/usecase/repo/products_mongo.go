package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/qulaz/artforintrovert-test/internal/common/commonerr"
	"github.com/qulaz/artforintrovert-test/internal/entity"
	"github.com/qulaz/artforintrovert-test/internal/tracing"
	"github.com/qulaz/artforintrovert-test/internal/types"
	"github.com/qulaz/artforintrovert-test/pkg/logging"
)

const (
	collectionName      = "products"
	notFoundMsgTemplate = "product with id %s not found"
)

type MongoRepository struct {
	collection *mongo.Collection
	logger     logging.ContextLogger
}

func NewMongoRepository(mongo *mongo.Database, logger logging.ContextLogger) *MongoRepository {
	return &MongoRepository{
		collection: mongo.Collection(collectionName),
		logger:     logger,
	}
}

func (r *MongoRepository) GetProducts(ctx context.Context) ([]*entity.Product, error) {
	ctx, span := tracing.Tracer.Start(ctx, "repository.GetProducts")
	defer span.End()

	var products []*entity.Product

	cursor, err := r.collection.Find(ctx, bson.D{}, options.Find().SetSort(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *MongoRepository) UpdateProduct(
	ctx context.Context,
	productId types.Id,
	updatedProduct *entity.Product,
) (*entity.Product, error) {
	ctx, span := tracing.Tracer.Start(ctx, "repository.UpdateProduct")
	defer span.End()

	res, err := r.collection.ReplaceOne(ctx, bson.M{"_id": productId}, updatedProduct)
	if err != nil {
		return nil, err
	}

	if res.MatchedCount == 0 {
		return nil, commonerr.NewNotFoundError(notFoundMsgTemplate, productId.Hex())
	}

	return updatedProduct, nil
}

func (r *MongoRepository) DeleteProduct(ctx context.Context, id types.Id) error {
	ctx, span := tracing.Tracer.Start(ctx, "repository.DeleteProduct")
	defer span.End()

	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return commonerr.NewNotFoundError(notFoundMsgTemplate, id.Hex())
	}

	return nil
}

func (r *MongoRepository) CreateProducts(ctx context.Context, products []*entity.Product) error {
	ctx, span := tracing.Tracer.Start(ctx, "repository.CreateProducts")
	defer span.End()

	newProducts := make([]interface{}, len(products))

	for i := range products {
		newProducts[i] = products[i]
	}

	_, err := r.collection.InsertMany(ctx, newProducts)
	if err != nil {
		return err
	}

	return nil
}
