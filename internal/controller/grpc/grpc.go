package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/qulaz/artforintrovert-test/gen/api/v1"
	"github.com/qulaz/artforintrovert-test/internal/common/commonerr"
	"github.com/qulaz/artforintrovert-test/internal/entity"
	"github.com/qulaz/artforintrovert-test/internal/entity/mapper"
	"github.com/qulaz/artforintrovert-test/internal/tracing"
	"github.com/qulaz/artforintrovert-test/internal/types"
	"github.com/qulaz/artforintrovert-test/internal/usecase"
	"github.com/qulaz/artforintrovert-test/pkg/logging"
)

var _ api.ProductServiceServer = (*ProductGrpcServer)(nil)

type ProductGrpcServer struct {
	api.UnimplementedProductServiceServer
	useCase usecase.Product
	logger  logging.ContextLogger
}

func NewProductGrpcServer(useCase usecase.Product, logger logging.ContextLogger) *ProductGrpcServer {
	return &ProductGrpcServer{
		UnimplementedProductServiceServer: api.UnimplementedProductServiceServer{},
		useCase:                           useCase,
		logger:                            logger,
	}
}

func (p *ProductGrpcServer) GetProducts(ctx context.Context, req *api.GetProductsRequest) (*api.ProductList, error) {
	ctx, _ = p.logger.FromContext(ctx)
	ctx, span := tracing.Tracer.Start(ctx, "ProductGrpcServer.GetProducts")
	defer span.End()

	products, err := p.useCase.GetProducts(ctx, uint(req.GetLimit()), uint(req.GetOffset()))
	if err != nil {
		return nil, commonerr.GrpcErrHandler(ctx, err, &commonerr.SentryInfo{
			Contexts: map[string]interface{}{"limit": req.Limit, "offset": req.Offset},
		})
	}

	return &api.ProductList{
		Products: mapper.ManyProductsToGrpc(products),
	}, nil
}

func (p *ProductGrpcServer) UpdateProduct(ctx context.Context, product *api.Product) (*api.Product, error) {
	ctx, _ = p.logger.FromContext(ctx)
	ctx, span := tracing.Tracer.Start(ctx, "ProductGrpcServer.UpdateProduct")
	defer span.End()

	sentryInfo := &commonerr.SentryInfo{
		Contexts: map[string]interface{}{"product": product},
	}

	productId, err := types.NewIdFromString(product.Id)
	if err != nil {
		return nil, commonerr.GrpcErrHandler(ctx, err, sentryInfo)
	}

	updatedProduct, err := p.useCase.UpdateProduct(ctx, productId, &entity.Product{
		Id:          productId,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
	})
	if err != nil {
		return nil, commonerr.GrpcErrHandler(ctx, err, sentryInfo)
	}

	return &api.Product{
		Id:          updatedProduct.Id.Hex(),
		Name:        updatedProduct.Name,
		Description: updatedProduct.Description,
		Price:       updatedProduct.Price,
	}, nil
}

func (p *ProductGrpcServer) DeleteProduct(ctx context.Context, id *api.Id) (*emptypb.Empty, error) {
	ctx, _ = p.logger.FromContext(ctx)
	ctx, span := tracing.Tracer.Start(ctx, "ProductGrpcServer.UpdateProduct")
	defer span.End()

	sentryInfo := &commonerr.SentryInfo{
		Contexts: map[string]interface{}{"productId": id.Id},
	}

	productId, err := types.NewIdFromString(id.Id)
	if err != nil {
		return nil, commonerr.GrpcErrHandler(ctx, err, sentryInfo)
	}

	if err := p.useCase.DeleteProduct(ctx, productId); err != nil {
		return nil, commonerr.GrpcErrHandler(ctx, err, sentryInfo)
	}

	return &emptypb.Empty{}, nil
}
