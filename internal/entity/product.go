package entity

import (
	"github.com/qulaz/artforintrovert-test/internal/common/commonerr"
	"github.com/qulaz/artforintrovert-test/internal/types"
)

type Product struct {
	Id          types.Id `json:"id" bson:"_id"`
	Name        string   `json:"name" bson:"name"`
	Description string   `json:"description" bson:"description"`
	Price       int32    `json:"price" bson:"price"`
}

func (p *Product) Hash() string {
	return p.Id.Hex()
}

func (p *Product) Validate() error {
	if p.Name == "" {
		return commonerr.NewIncorrectInputError("name is required")
	}

	if p.Description == "" {
		return commonerr.NewIncorrectInputError("description is required")
	}

	if p.Price <= 0 {
		return commonerr.NewIncorrectInputError("price must be greater than 0")
	}

	return nil
}
