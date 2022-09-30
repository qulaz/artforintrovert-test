package mapper

import (
	"github.com/qulaz/artforintrovert-test/gen/api/v1"
	"github.com/qulaz/artforintrovert-test/internal/entity"
	"github.com/qulaz/artforintrovert-test/internal/types"
)

func OneProductToGrpc(product *entity.Product) *api.Product {
	return &api.Product{
		Id:          product.Id.Hex(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
	}
}

func ManyProductsToGrpc(products []*entity.Product) []*api.Product {
	grpcProducts := make([]*api.Product, len(products))

	for i, product := range products {
		grpcProducts[i] = OneProductToGrpc(product)
	}

	return grpcProducts
}

func OneGrpcProductToEntity(product *api.Product) (*entity.Product, error) {
	id, err := types.NewIdFromString(product.Id)
	if err != nil {
		return nil, err
	}

	return &entity.Product{
		Id:          id,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
	}, nil
}
