package cache

import (
	"github.com/puzpuzpuz/xsync/v3"
)

type Cache struct {
	xmap *xsync.Map
}

func NewCache() *Cache {
	return &Cache{
		xmap: xsync.NewMap(),
	}
}

func (c *Cache) Purge() error {
	c.xmap.Clear()
	return nil
}

func (c *Cache) Exists(key string) bool {
	_, ok := c.xmap.Load(key)
	return ok
}

func (c *Cache) Add(key string) {
	c.xmap.Store(key, nil)
}

func (c *Cache) Remove(key string) {
	c.xmap.Delete(key)
}

func (c *Cache) Size() int {
	return c.xmap.Size()
}
