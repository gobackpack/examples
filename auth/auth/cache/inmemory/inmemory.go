package inmemory

import (
	"encoding/json"
	"github.com/gobackpack/examples/auth/auth/cache"
	"github.com/muesli/cache2go"
)

type Cache struct {
	Engine *cache2go.CacheTable
}

func New(tableName string) *Cache {
	return &Cache{
		Engine: cache2go.Cache(tableName),
	}
}

func (c *Cache) Store(items ...*cache.Item) error {
	for _, item := range items {
		c.Engine.Add(item.Key, item.Expiration, item.Value)
	}

	return nil
}

func (c *Cache) Get(keys ...string) ([]byte, error) {
	var result []byte

	for _, k := range keys {
		item, err := c.Engine.Value(k)
		if err != nil {
			continue
		}

		bItem, err := json.Marshal(item.Data())
		if err != nil {
			continue
		}

		result = append(result, bItem...)
	}

	return result, nil
}

func (c *Cache) Delete(keys ...string) error {
	for k := range keys {
		c.Engine.Delete(k)
	}

	return nil
}
