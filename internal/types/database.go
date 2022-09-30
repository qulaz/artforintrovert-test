package types

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/qulaz/artforintrovert-test/internal/common/commonerr"
)

type Id = primitive.ObjectID

func NewIdFromString(rawId string) (Id, error) {
	objectId, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		return primitive.ObjectID{}, commonerr.NewIncorrectInputError("wrong id format")
	}

	return objectId, nil
}

func NewId() Id {
	return primitive.NewObjectID()
}
