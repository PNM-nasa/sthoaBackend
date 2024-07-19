package apiquestion

import "go.mongodb.org/mongo-driver/mongo"

func Setup(collection *mongo.Collection) {
	Collection = collection
}
