package repository

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func optionsSortUpdated() *options.FindOptions {
	return options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}})
}
