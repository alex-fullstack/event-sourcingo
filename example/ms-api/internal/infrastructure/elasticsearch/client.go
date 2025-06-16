package elasticsearch

import (
	"api/internal/domain/dto"
	"api/internal/domain/entities"
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

type Client struct {
	db *elasticsearch.TypedClient
}

func NewClient(cfg elasticsearch.Config) (*Client, error) {
	client, err := elasticsearch.NewTypedClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{db: client}, nil
}

func (c *Client) Get(ctx context.Context, ids []string) ([]*entities.User, error) {
	users := make([]*entities.User, 0)
	exists, err := c.db.Indices.Exists("users").Do(ctx)
	if err != nil {
		return users, err
	}

	if !exists {
		return users, nil
	}
	var query *types.Query
	if len(ids) > 0 {
		query = &types.Query{
			Terms: &types.TermsQuery{
				TermsQuery: map[string]types.TermsQueryField{
					"_id": types.TermsQueryField(ids),
				},
			},
		}
	} else {
		query = &types.Query{
			MatchAll: &types.MatchAllQuery{},
		}
	}
	req := &search.Request{
		Query: query,
	}
	r, err := c.db.Search().
		Index("users").
		Request(req).Do(context.Background())
	if err != nil {
		return users, err
	}
	for _, hit := range r.Hits.Hits {
		var user dto.UserDocument
		err = json.Unmarshal(hit.Source_, &user)
		if err != nil {
			return users, err
		}
		users = append(users, entities.NewUser(user.ID, user.Info, user.Policies))
	}

	return users, nil
}

func (c *Client) Upsert(ctx context.Context, user *entities.User) error {
	doc := user.Document()
	_, err := c.db.Index("users").Id(doc.ID.String()).Request(doc).Do(ctx)
	return err
}
