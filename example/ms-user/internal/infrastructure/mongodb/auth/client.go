package auth

import (
	"context"
	"errors"
	"fmt"
	"user/internal/domain/dto"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	collectionName       = "user"
	databaseName         = "projections"
	bsonDateTimeTypeName = "bson.DateTime"
)

type MongoDB struct {
	conn *mongo.Client
}

func NewMongoDB(ctx context.Context, config *options.ClientOptions) (*MongoDB, error) {
	client, err := mongo.Connect(config)
	if err != nil {
		return nil, err
	}
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	if err = migrate(ctx, client); err != nil {
		return nil, err
	}

	return &MongoDB{conn: client}, nil
}

func (db *MongoDB) Get(ctx context.Context, id uuid.UUID) (dto.UserProjection, error) {
	var document bson.M
	if err := db.conn.Database(databaseName).Collection(collectionName).FindOne(
		ctx,
		bson.M{"id": id.String()},
	).Decode(&document); err != nil {
		return dto.UserProjection{}, err
	}
	return db.parseDocument(document)
}

func (db *MongoDB) GetByEmail(ctx context.Context, email string) (dto.UserProjection, error) {
	var document bson.M
	if err := db.conn.Database(databaseName).Collection(collectionName).FindOne(
		ctx,
		bson.M{"credentials.email": email},
	).Decode(&document); err != nil {
		return dto.UserProjection{}, err
	}
	return db.parseDocument(document)
}
func (db *MongoDB) Save(ctx context.Context, data interface{}) error {
	opts := options.FindOneAndUpdate().SetUpsert(true)
	var updatedDocument bson.M
	checkedData, ok := data.(dto.UserProjection)
	if !ok {
		return errors.New("wrong data type, need dto.UserProjection")
	}
	err := db.conn.Database(databaseName).Collection(collectionName).FindOneAndUpdate(
		ctx,
		bson.M{"id": checkedData.ID.String()},
		bson.M{"$set": bson.M{
			"id":           checkedData.ID.String(),
			"credentials":  checkedData.Credentials,
			"confirmation": checkedData.Confirmation,
			"history":      checkedData.History,
		}},
		opts,
	).Decode(&updatedDocument)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	return nil
}

func (db *MongoDB) Close(ctx context.Context) error {
	return db.conn.Disconnect(ctx)
}

func (db *MongoDB) parseDocument(document bson.M) (dto.UserProjection, error) {
	idField, ok := document["id"].(string)
	if !ok {
		return dto.UserProjection{}, errors.New("wrong data type, need string")
	}
	id, err := uuid.Parse(idField)
	if err != nil {
		return dto.UserProjection{}, err
	}
	history, err := db.parseHistory(document)
	if err != nil {
		return dto.UserProjection{}, err
	}
	credentials, err := db.parseCredentials(document)
	if err != nil {
		return dto.UserProjection{}, err
	}
	confirmation, err := db.parseConfirmation(document)
	if err != nil {
		return dto.UserProjection{}, err
	}
	return dto.NewUserProjection(id, credentials, confirmation, history), nil
}

func (db *MongoDB) parseConfirmation(document bson.M) (*dto.ConfirmationProjection, error) {
	var confirmation *dto.ConfirmationProjection
	rawConfirmation := document["confirmation"]
	if rawConfirmation != nil {
		confirmationField, ok := rawConfirmation.(bson.D)
		if !ok {
			return nil, errors.New("wrong data type, need bson.D")
		}
		confirmation = &dto.ConfirmationProjection{}
		for _, item := range confirmationField {
			var expired bson.DateTime
			var needType = "string"
			switch item.Key {
			case "code":
				confirmation.Code, ok = item.Value.(string)
			case "expired":
				expired, ok = item.Value.(bson.DateTime)
				needType = bsonDateTimeTypeName
			}
			if !ok {
				return nil, fmt.Errorf("wrong data type, need %s", needType)
			}
			if needType == bsonDateTimeTypeName {
				confirmation.Expired = expired.Time()
			}
		}
	}
	return confirmation, nil
}

func (db *MongoDB) parseCredentials(document bson.M) (*dto.CredentialsProjection, error) {
	var credentials *dto.CredentialsProjection
	rawCredentials := document["credentials"]
	if rawCredentials != nil {
		credentials = &dto.CredentialsProjection{}
		credentialsField, ok := rawCredentials.(bson.D)
		if !ok {
			return nil, errors.New("wrong data type, need bson.D")
		}
		for _, item := range credentialsField {
			switch item.Key {
			case "email":
				credentials.Email, ok = item.Value.(string)
			case "password_hash":
				credentials.PasswordHash, ok = item.Value.(string)
			}
			if !ok {
				return nil, errors.New("wrong data type, need string")
			}
		}
	}
	return credentials, nil
}

func (db *MongoDB) parseHistory(document bson.M) ([]dto.ActivityRecord, error) {
	historyField, ok := document["history"].(bson.A)
	if !ok {
		return nil, errors.New("wrong data type, need bson.A")
	}
	history := make([]dto.ActivityRecord, len(historyField))
	for i, record := range historyField {
		var activity = dto.ActivityRecord{}
		recordField, reqOk := record.(bson.D)
		if !reqOk {
			return nil, errors.New("wrong data type, need bson.D")
		}
		for _, item := range recordField {
			var timestamp bson.DateTime
			var needType = "string"
			switch item.Key {
			case "type":
				activity.Type, ok = item.Value.(string)
			case "device":
				activity.Device, ok = item.Value.(string)
			case "timestamp":
				timestamp, ok = item.Value.(bson.DateTime)
				needType = bsonDateTimeTypeName
			}
			if !ok {
				return nil, fmt.Errorf("wrong data type, need %s", needType)
			}
			if needType == bsonDateTimeTypeName {
				activity.Timestamp = timestamp.Time()
			}
		}
		history[i] = activity
	}
	return history, nil
}
