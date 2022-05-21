package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetDBCollection(collectionName string) (*mongo.Collection, error) {
	// Connect to database
	client, err := mongo.Connect(context.TODO(), options.Client())

	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	collection := client.Database("ScheduleGenerator").Collection(collectionName)
	return collection, nil
}
