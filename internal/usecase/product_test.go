package usecase

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qulaz/artforintrovert-test/internal/common/commonerr"
	"github.com/qulaz/artforintrovert-test/internal/entity"
	"github.com/qulaz/artforintrovert-test/internal/types"
	"github.com/qulaz/artforintrovert-test/internal/usecase/repo"
	"github.com/qulaz/artforintrovert-test/pkg/cache"
	"github.com/qulaz/artforintrovert-test/pkg/logging"
)

func newProductUseCase(t *testing.T) (
	*ProductUseCase,
	*repo.MockRepository,
	*cache.MockEntityCache[*entity.Product],
	func(),
) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockRepo := repo.NewMockRepository(ctrl)
	mockCache := cache.NewMockEntityCache[*entity.Product](ctrl)

	teardown := func() {
		ctrl.Finish()
	}

	return &ProductUseCase{
		repo:     mockRepo,
		logger:   logging.NewDummyLogger(),
		cache:    mockCache,
		cacheTtl: 10,
	}, mockRepo, mockCache, teardown
}

func TestProductUseCase_GetProducts(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		uc, _, mockCache, teardown := newProductUseCase(t)
		defer teardown()

		mockCache.EXPECT().GetList(uint(100), uint(0)).Return([]*entity.Product{}, nil)

		products, err := uc.GetProducts(context.Background(), 100, 0)
		require.NoError(t, err)
		assert.Equal(t, []*entity.Product{}, products)
	})
	t.Run("zero limit", func(t *testing.T) {
		t.Parallel()
		uc, _, mockCache, teardown := newProductUseCase(t)
		defer teardown()

		mockCache.EXPECT().GetList(uint(defaultLimit), uint(0)).Return([]*entity.Product{}, nil)

		products, err := uc.GetProducts(context.Background(), 0, 0)
		require.NoError(t, err)
		assert.Equal(t, []*entity.Product{}, products)
	})
	t.Run("cache error", func(t *testing.T) {
		t.Parallel()
		uc, _, mockCache, teardown := newProductUseCase(t)
		defer teardown()

		mockCache.EXPECT().GetList(uint(100), uint(0)).Return(nil, errors.New(""))

		products, err := uc.GetProducts(context.Background(), 100, 0)
		require.Error(t, err)
		assert.Nil(t, products)
	})
}

func TestProductUseCase_UpdateProduct(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		uc, mockRepo, mockCache, teardown := newProductUseCase(t)
		defer teardown()

		product := newValidProduct()
		updatedProduct := newValidProduct()

		mockRepo.EXPECT().UpdateProduct(gomock.Any(), product.Id, updatedProduct).Return(updatedProduct, nil)
		mockCache.EXPECT().Set(updatedProduct).Return(nil)

		res, err := uc.UpdateProduct(context.Background(), product.Id, updatedProduct)
		require.NoError(t, err)
		require.Equal(t, res, updatedProduct)
	})
	t.Run("Invalid product", func(t *testing.T) {
		t.Parallel()
		uc, _, _, teardown := newProductUseCase(t)
		defer teardown()

		product := newValidProduct()
		updatedProduct := newInvalidProduct()
		var appError commonerr.AppError

		res, err := uc.UpdateProduct(context.Background(), product.Id, updatedProduct)
		require.Error(t, err)
		require.True(t, commonerr.IsAppError(err))
		require.True(t, errors.As(err, &appError), err)
		assert.Equal(t, commonerr.ErrorTypeIncorrectInput, appError.ErrorType())
		require.Nil(t, res)
	})
	t.Run("product not found", func(t *testing.T) {
		t.Parallel()
		uc, mockRepo, _, teardown := newProductUseCase(t)
		defer teardown()

		product := newValidProduct()
		updatedProduct := newValidProduct()
		notFoundErr := commonerr.NewNotFoundError("product %s not found", product.Id.Hex())
		var appError commonerr.AppError

		mockRepo.EXPECT().UpdateProduct(gomock.Any(), product.Id, updatedProduct).Return(nil, notFoundErr)

		res, err := uc.UpdateProduct(context.Background(), product.Id, updatedProduct)
		require.Error(t, err)
		require.True(t, commonerr.IsAppError(err))
		require.True(t, errors.As(err, &appError))
		require.Equal(t, commonerr.ErrorTypeNotFound, appError.ErrorType())
		require.Nil(t, res)
	})
	t.Run("Cache error", func(t *testing.T) {
		t.Parallel()
		uc, mockRepo, cacheMock, teardown := newProductUseCase(t)
		defer teardown()

		product := newValidProduct()
		updatedProduct := newValidProduct()

		mockRepo.EXPECT().UpdateProduct(gomock.Any(), product.Id, updatedProduct).Return(updatedProduct, nil)
		cacheMock.EXPECT().Set(updatedProduct).Return(errors.New(""))

		res, err := uc.UpdateProduct(context.Background(), product.Id, updatedProduct)
		require.NoError(t, err)
		require.Equal(t, updatedProduct, res)
	})
}

func TestProductUseCase_DeleteProduct(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		product := newValidProduct()
		uc, mockRepo, mockCache, teardown := newProductUseCase(t)
		defer teardown()

		mockRepo.EXPECT().DeleteProduct(gomock.Any(), product.Id).Return(nil)
		mockCache.EXPECT().Delete(product.Id.Hex()).Return(nil)

		err := uc.DeleteProduct(context.Background(), product.Id)
		require.NoError(t, err)
	})
	t.Run("product not found", func(t *testing.T) {
		t.Parallel()
		product := newValidProduct()
		uc, mockRepo, _, teardown := newProductUseCase(t)
		defer teardown()

		notFoundErr := commonerr.NewNotFoundError("product %s not found", product.Id.Hex())
		var appError commonerr.AppError

		mockRepo.EXPECT().DeleteProduct(gomock.Any(), product.Id).Return(notFoundErr)

		err := uc.DeleteProduct(context.Background(), product.Id)
		require.Error(t, err)
		require.True(t, commonerr.IsAppError(err))
		require.True(t, errors.As(err, &appError))
		require.Equal(t, commonerr.ErrorTypeNotFound, appError.ErrorType())
	})
	t.Run("Cache error", func(t *testing.T) {
		t.Parallel()
		product := newValidProduct()
		uc, mockRepo, cacheMock, teardown := newProductUseCase(t)
		defer teardown()

		mockRepo.EXPECT().DeleteProduct(gomock.Any(), product.Id).Return(nil)
		cacheMock.EXPECT().Delete(product.Id.Hex()).Return(errors.New(""))

		err := uc.DeleteProduct(context.Background(), product.Id)
		require.NoError(t, err)
	})
}

func TestProductUseCase_syncCache(t *testing.T) {
	products := []*entity.Product{newValidProduct(), newValidProduct(), newValidProduct()}

	testCases := []struct {
		name string
		mock func(mockRepo *repo.MockRepository, mockCache *cache.MockEntityCache[*entity.Product])
	}{
		{
			name: "Success",
			mock: func(mockRepo *repo.MockRepository, mockCache *cache.MockEntityCache[*entity.Product]) {
				mockRepo.EXPECT().GetProducts(gomock.Any()).Return(products, nil)
				mockCache.EXPECT().Replace(products).Return(nil)
			},
		},
		{
			name: "Repo error",
			mock: func(mockRepo *repo.MockRepository, mockCache *cache.MockEntityCache[*entity.Product]) {
				mockRepo.EXPECT().GetProducts(gomock.Any()).Return(nil, errors.New(""))
			},
		},
		{
			name: "Cache error",
			mock: func(mockRepo *repo.MockRepository, mockCache *cache.MockEntityCache[*entity.Product]) {
				mockRepo.EXPECT().GetProducts(gomock.Any()).Return(products, nil)
				mockCache.EXPECT().Replace(products).Return(errors.New(""))
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uc, mockRepo, mockCache, teardown := newProductUseCase(t)
			defer teardown()

			// test fails if has been called unexpected func or expected not been called
			tc.mock(mockRepo, mockCache)
			uc.syncCache(context.Background())
		})
	}
}

func newValidProduct() *entity.Product {
	return &entity.Product{
		Id:          types.NewId(),
		Name:        gofakeit.Name(),
		Description: gofakeit.JobDescriptor(),
		Price:       int32(gofakeit.IntRange(1, math.MaxInt32)),
	}
}

func newInvalidProduct() *entity.Product {
	return &entity.Product{
		Id:          types.NewId(),
		Name:        gofakeit.Name(),
		Description: gofakeit.JobDescriptor(),
		Price:       -1,
	}
}
