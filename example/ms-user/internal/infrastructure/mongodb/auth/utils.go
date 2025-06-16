package auth

import (
	"context"
	"fmt"
	"slices"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const emailUniqueIndexValue = 1

func migrate(ctx context.Context, client *mongo.Client) error {
	indexes := client.Database(databaseName).Collection(collectionName).Indexes()
	cursor, err := indexes.List(ctx)
	if err != nil {
		return err
	}
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return err
	}
	hasNoEmailUniqIndex := slices.IndexFunc(results, func(m bson.M) bool {
		return m["name"] == fmt.Sprintf("email_%d", emailUniqueIndexValue)
	}) == -1
	if hasNoEmailUniqIndex {
		_, err = indexes.CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "credentials.email", Value: emailUniqueIndexValue}},
			Options: options.Index().SetUnique(true),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
