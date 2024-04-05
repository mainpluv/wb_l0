package cache

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/mainpluv/wb_l0/internal/model"
)

type Cache interface {
	Get(uuid.UUID) (*model.Order, error)
	Put(*model.Order)
}

type MemoryCache struct {
	cache sync.Map
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{}
}

func (c *MemoryCache) Get(uuid uuid.UUID) (*model.Order, error) {
	order, ok := c.cache.Load(uuid)
	res, ok1 := order.(*model.Order)
	if !ok1 {
		return nil, fmt.Errorf("Error in cache")
	}
	if !ok {
		return nil, fmt.Errorf("Order not found")
	}
	return res, nil
}

func (c *MemoryCache) Put(order *model.Order) {
	c.cache.Store(order.OrderUUID, order)
	fmt.Printf("saved new order in cash with UUID: %s\n", order.OrderUUID.String())
}
