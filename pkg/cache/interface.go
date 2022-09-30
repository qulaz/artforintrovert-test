package cache

import "errors"

var ErrKeyNotFound = errors.New("key not found")

type Hashable interface {
	Hash() string
}

//go:generate go run github.com/golang/mock/mockgen -source=interface.go -destination=mock.go -package=cache
type EntityCache[V Hashable] interface {
	Get(key string) (V, error)
	Set(value V) error
	Delete(key string) error
	GetList(limit uint, offset uint) ([]V, error)
	Replace(values []V) error
}
