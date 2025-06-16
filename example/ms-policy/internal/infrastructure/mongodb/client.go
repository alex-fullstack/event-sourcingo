package mongodb

import (
	"context"
	"errors"
	"policy/internal/domain/dto"
	"slices"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	collectionName = "policy"
	databaseName   = "projections"
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

func (db *MongoDB) Get(ctx context.Context, id uuid.UUID) (*dto.PolicyProjection, error) {
	var document bson.M
	if err := db.conn.Database(databaseName).Collection(collectionName).FindOne(
		ctx,
		bson.M{"id": id.String()},
	).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &dto.PolicyProjection{}, dto.ErrEmptyResult
		}
		return &dto.PolicyProjection{}, err
	}
	res, err := db.parseDocument(document)
	if err != nil {
		return &dto.PolicyProjection{}, err
	}
	return &res, nil
}

func (db *MongoDB) GetByRoleCode(ctx context.Context, code string) (*dto.PolicyProjection, error) {
	var document bson.M
	if err := db.conn.Database(databaseName).Collection(collectionName).FindOne(
		ctx,
		bson.M{"role.code": code},
	).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &dto.PolicyProjection{}, dto.ErrEmptyResult
		}
		return &dto.PolicyProjection{}, err
	}
	res, err := db.parseDocument(document)
	if err != nil {
		return &dto.PolicyProjection{}, err
	}
	return &res, nil
}

func (db *MongoDB) GetByUserID(ctx context.Context, id uuid.UUID) ([]dto.PolicyProjection, error) {
	cur, err := db.conn.Database(databaseName).Collection(collectionName).Find(ctx, bson.M{"users": id.String()})
	if err != nil {
		return nil, err
	}
	var projections []dto.PolicyProjection
	defer func() {
		err = cur.Close(ctx)
		if err != nil {
			projections = nil
		}
	}()

	for cur.Next(ctx) {
		var projection dto.PolicyProjection
		var find bson.M
		err = cur.Decode(&find)
		if err != nil {
			return nil, err
		}
		projection, err = db.parseDocument(find)
		if err != nil {
			return nil, err
		}
		projections = append(projections, projection)
	}

	if err = cur.Err(); err != nil {
		return nil, err
	}
	return projections, nil
}

func (db *MongoDB) Save(ctx context.Context, data interface{}) error {
	opts := options.FindOneAndUpdate().SetUpsert(true)
	var updatedDocument bson.M
	checkedData, ok := data.(dto.PolicyProjection)
	if !ok {
		return errors.New("wrong data type, need dto.PolicyProjection")
	}
	err := db.conn.Database(databaseName).Collection(collectionName).FindOneAndUpdate(
		ctx,
		bson.M{"id": checkedData.ID.String()},
		bson.M{"$set": bson.M{
			"id":          checkedData.ID.String(),
			"role":        checkedData.Role,
			"permissions": checkedData.Permissions,
			"users": slices.Collect(func(yield func(string) bool) {
				for _, user := range checkedData.Users {
					if !yield(user.String()) {
						return
					}
				}
			}),
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

func (db *MongoDB) parseDocument(document bson.M) (dto.PolicyProjection, error) {
	idField, ok := document["id"].(string)
	if !ok {
		return dto.PolicyProjection{}, errors.New("wrong data type, need string")
	}
	id, err := uuid.Parse(idField)
	if err != nil {
		return dto.PolicyProjection{}, err
	}
	users, err := db.parseUsers(document)
	if err != nil {
		return dto.PolicyProjection{}, err
	}
	permissions, err := db.parsePermissions(document)
	if err != nil {
		return dto.PolicyProjection{}, err
	}
	role, err := db.parseRole(document)
	if err != nil {
		return dto.PolicyProjection{}, err
	}
	return dto.NewPolicyProjection(id, role, permissions, users), nil
}

func (db *MongoDB) parseRole(document bson.M) (*dto.RoleProjection, error) {
	rawRole := document["role"]
	var role *dto.RoleProjection
	if rawRole != nil {
		rawRoleTyped, ok := rawRole.(bson.D)
		if !ok {
			return nil, errors.New("wrong data type, need bson.D")
		}
		role = &dto.RoleProjection{}
		for _, item := range rawRoleTyped {
			switch item.Key {
			case "code":
				role.Code, ok = item.Value.(string)
			case "name":
				role.Name, ok = item.Value.(string)
			}
			if !ok {
				return nil, errors.New("wrong data type, need string")
			}
		}
	}
	return role, nil
}

func (db *MongoDB) parsePermissions(document bson.M) ([]dto.PermissionProjection, error) {
	rawPermissions := document["permissions"]
	if rawPermissions != nil {
		permissionsField, ok := document["permissions"].(bson.A)
		if !ok {
			return nil, errors.New("wrong data type, need bson.A")
		}
		permissions := make([]dto.PermissionProjection, len(permissionsField))
		for i, record := range permissionsField {
			var perm = dto.PermissionProjection{}
			recordField, reqOk := record.(bson.D)
			if !reqOk {
				return nil, errors.New("wrong data type, need bson.D")
			}
			for _, item := range recordField {
				switch item.Key {
				case "code":
					perm.Code, ok = item.Value.(string)
				case "name":
					perm.Name, ok = item.Value.(string)
				}
				if !ok {
					return nil, errors.New("wrong data type, need string")
				}
			}
			permissions[i] = perm
		}
		return permissions, nil
	}

	return []dto.PermissionProjection{}, nil
}

func (db *MongoDB) parseUsers(document bson.M) ([]uuid.UUID, error) {
	rawUsers := document["users"]
	if rawUsers != nil {
		usersField, ok := document["users"].(bson.A)
		if !ok {
			return nil, errors.New("wrong data type, need bson.A")
		}
		users := make([]uuid.UUID, len(usersField))
		for i, record := range usersField {
			recordField, reqOk := record.(string)
			if !reqOk {
				return nil, errors.New("wrong data type, need string")
			}
			var err error
			users[i], err = uuid.Parse(recordField)
			if err != nil {
				return nil, err
			}
		}
		return users, nil
	}
	return []uuid.UUID{}, nil
}
