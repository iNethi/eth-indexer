package cache

import "github.com/puzpuzpuz/xsync/v3"

type Cache struct {
	provider *xsync.MapOf[string, bool]
}

func New() *Cache {
	return &Cache{
		provider: xsync.NewMapOf[string, bool](),
	}
}

func (c *Cache) Set(key string) {
	c.provider.Store(key, true)
}

func (c *Cache) Get(key string) bool {
	v, _ := c.provider.Load(key)
	return v
}

func (c *Cache) Size() int {
	return c.provider.Size()
}
