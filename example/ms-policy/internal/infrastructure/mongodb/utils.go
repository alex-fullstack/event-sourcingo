package mongodb

import (
	"context"
	"fmt"
	"slices"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const uniqueIndexValue = 1

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
	hasNoRoleCodeUniqIndex := slices.IndexFunc(results, func(m bson.M) bool {
		return m["name"] == fmt.Sprintf("role.code_%d", uniqueIndexValue)
	}) == -1
	if hasNoRoleCodeUniqIndex {
		_, err = indexes.CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "role.code", Value: uniqueIndexValue}},
			Options: options.Index().SetUnique(true),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
