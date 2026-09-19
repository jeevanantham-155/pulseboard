package repository

import "go.mongodb.org/mongo-driver/mongo/options"

func optionsUnique() *options.IndexOptions {
	unique := true
	return &options.IndexOptions{Unique: &unique}
}
