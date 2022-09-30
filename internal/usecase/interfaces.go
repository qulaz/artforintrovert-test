package usecase

import (
	"context"

	"github.com/qulaz/artforintrovert-test/internal/entity"
	"github.com/qulaz/artforintrovert-test/internal/types"
)

//go:generate go run github.com/golang/mock/mockgen -source=interfaces.go -destination=product_mock.go -package=usecase
type Product interface {
	GetProducts(ctx context.Context, limit uint, offset uint) ([]*entity.Product, error)
	UpdateProduct(ctx context.Context, productId types.Id, updatedProduct *entity.Product) (*entity.Product, error)
	DeleteProduct(ctx context.Context, id types.Id) error
}

//go:generate go run github.com/golang/mock/mockgen -source=interfaces.go -destination=repo/products_mock.go -package=repo
type Repository interface {
	GetProducts(ctx context.Context) ([]*entity.Product, error)
	UpdateProduct(ctx context.Context, productId types.Id, updatedProduct *entity.Product) (*entity.Product, error)
	DeleteProduct(ctx context.Context, id types.Id) error
	CreateProducts(ctx context.Context, product []*entity.Product) error
}
