package cache

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/mainpluv/wb_l0/internal/model"
)

type Cache interface {
	Get(uuid.UUID) (*model.Order, error)
	Put(model.Order)
}

type MemoryCache struct {
	cache sync.Map
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{}
}

func (c *MemoryCache) Get(uuid uuid.UUID) (*model.Order, error) {
	order, ok := c.cache.Load(uuid)
	if !ok {
		return nil, fmt.Errorf("Order not found")
	}
	return order.(*model.Order), nil
}

func (c *MemoryCache) Put(order model.Order) {
	c.cache.Store(order.OrderUUID, order)
}
