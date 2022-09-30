package usecase

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/qulaz/artforintrovert-test/internal/common/commonerr"
	"github.com/qulaz/artforintrovert-test/internal/entity"
	"github.com/qulaz/artforintrovert-test/internal/tracing"
	"github.com/qulaz/artforintrovert-test/internal/types"
	"github.com/qulaz/artforintrovert-test/pkg/cache"
	"github.com/qulaz/artforintrovert-test/pkg/logging"
)

const (
	defaultLimit = 100
)

var _ Product = (*ProductUseCase)(nil)

type ProductUseCase struct {
	repo     Repository
	logger   logging.ContextLogger
	cache    cache.EntityCache[*entity.Product]
	cacheTtl time.Duration
}

func NewProductUseCase(
	repo Repository,
	cache cache.EntityCache[*entity.Product],
	logger logging.ContextLogger,
	cacheTtl time.Duration,
) *ProductUseCase {
	return &ProductUseCase{
		repo:     repo,
		logger:   logger,
		cache:    cache,
		cacheTtl: cacheTtl,
	}
}

func (p *ProductUseCase) GetProducts(ctx context.Context, limit uint, offset uint) ([]*entity.Product, error) {
	ctx, _ = p.logger.FromContext(ctx)
	_, span := tracing.Tracer.Start(ctx, "productUseCase.GetProducts")
	defer span.End()

	if limit == 0 {
		limit = defaultLimit
	}

	cachedProducts, err := p.cache.GetList(limit, offset)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return cachedProducts, nil
}

func (p *ProductUseCase) UpdateProduct(
	ctx context.Context,
	productId types.Id,
	updatedProduct *entity.Product,
) (*entity.Product, error) {
	ctx, logger := p.logger.FromContext(ctx, "productId", productId, "updatedProduct", updatedProduct)
	ctx, span := tracing.Tracer.Start(ctx, "productUseCase.UpdateProduct")
	defer span.End()

	if err := updatedProduct.Validate(); err != nil {
		return nil, err
	}

	product, err := p.repo.UpdateProduct(ctx, productId, updatedProduct)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := p.cache.Set(product); err != nil {
		commonerr.SendToSentry(ctx, errors.WithStack(err), &commonerr.SentryInfo{
			Contexts: map[string]interface{}{
				"updatedProduct": updatedProduct,
				"productId":      productId,
			},
		})
		logger.Warnw("error while updating product in cache", "err", err)
	}

	return product, nil
}

func (p *ProductUseCase) DeleteProduct(ctx context.Context, id types.Id) error {
	ctx, logger := p.logger.FromContext(ctx, "productId", id)
	ctx, span := tracing.Tracer.Start(ctx, "productUseCase.DeleteProduct")
	defer span.End()

	if err := p.repo.DeleteProduct(ctx, id); err != nil {
		return errors.WithStack(err)
	}

	if err := p.cache.Delete(id.Hex()); err != nil {
		commonerr.SendToSentry(ctx, errors.WithStack(err), &commonerr.SentryInfo{
			Contexts: map[string]interface{}{"productId": id},
		})
		logger.Warnw("error while deleting product from cache", "err", err)
	}

	return nil
}

func (p *ProductUseCase) SyncCache(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			p.syncCache(ctx)
			time.Sleep(p.cacheTtl)
		}
	}
}

func (p *ProductUseCase) syncCache(ctx context.Context) {
	ctx, span := tracing.Tracer.Start(ctx, "productUseCase.syncCache")
	defer span.End()

	p.logger.Infow("Start syncing cache")

	l, err := p.repo.GetProducts(ctx)
	if err != nil {
		commonerr.SendToSentry(ctx, errors.WithStack(err), nil)
		p.logger.Errorw("can't sync cache", "err", err)
		return
	}

	err = p.cache.Replace(l)
	if err != nil {
		commonerr.SendToSentry(ctx, errors.WithStack(err), nil)
		p.logger.Errorw("can't set batch in cache", "err", err)
		return
	}

	p.logger.Infow("Cache synced with database")
}
