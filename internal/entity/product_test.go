package entity

import (
	"errors"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qulaz/artforintrovert-test/internal/common/commonerr"
	"github.com/qulaz/artforintrovert-test/internal/types"
)

func isAppError(t *testing.T, err error) {
	t.Helper()

	var (
		appError1 commonerr.AppError
		appError2 commonerr.AppError
		appError3 commonerr.AppError

		wrappedErr  = fmt.Errorf("new context: %w", err)
		wrappedErr2 = fmt.Errorf("new context: %w", wrappedErr)
	)

	require.True(t, commonerr.IsAppError(err))
	require.True(t, commonerr.IsAppError(wrappedErr))
	require.True(t, commonerr.IsAppError(wrappedErr2))

	require.True(t, errors.As(err, &appError1))
	require.True(t, errors.As(wrappedErr, &appError2))
	require.True(t, errors.As(wrappedErr2, &appError3))

	require.True(t, errors.Is(err, wrappedErr2))

	assert.Equal(t, appError1.ErrorType(), commonerr.ErrorTypeIncorrectInput)
	assert.Equal(t, appError2.ErrorType(), commonerr.ErrorTypeIncorrectInput)
	assert.Equal(t, appError3.ErrorType(), commonerr.ErrorTypeIncorrectInput)
}

func TestProduct_Validate(t *testing.T) {
	t.Run("name required", func(t *testing.T) {
		product := &Product{ //nolint: exhaustruct
			Id:          types.NewId(),
			Description: gofakeit.Name(),
			Price:       1,
		}
		err := product.Validate()
		require.Error(t, err)
		isAppError(t, err)
	})
	t.Run("description required", func(t *testing.T) {
		product := &Product{ //nolint: exhaustruct
			Id:    types.NewId(),
			Name:  gofakeit.Name(),
			Price: 1,
		}
		err := product.Validate()
		require.Error(t, err)
		isAppError(t, err)
	})
	t.Run("price lower than 0", func(t *testing.T) {
		product := &Product{ //nolint: exhaustruct
			Id:    types.NewId(),
			Name:  gofakeit.Name(),
			Price: -1,
		}
		err := product.Validate()
		require.Error(t, err)
		isAppError(t, err)
	})
}
