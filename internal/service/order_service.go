package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mainpluv/wb_l0/internal/cache"
	"github.com/mainpluv/wb_l0/internal/database"
	"github.com/mainpluv/wb_l0/internal/model"
)

type OrderService interface {
	SaveOrder(order model.Order) (model.Order, error)
	GetOrder(uuid.UUID) (*model.Order, error)
	Pull() error
}

type OrderServiceImpl struct {
	repos database.OrderRepository
	cache cache.Cache
}

func NewOrderService(repos database.OrderRepository, cache cache.Cache) *OrderServiceImpl {
	return &OrderServiceImpl{
		repos: repos,
		cache: cache,
	}
}

func (s *OrderServiceImpl) SaveOrder(order model.Order) (model.Order, error) {
	// Реализация создания заказа
	newOrder, err := s.repos.Create(order)
	if err != nil {
		return model.Order{}, err
	}
	s.cache.Put(*newOrder)
	return *newOrder, nil
}

func (s *OrderServiceImpl) GetOrder(uuid uuid.UUID) (*model.Order, error) {
	// Реализация получения информации о заказе
	order, err := s.cache.Get(uuid)
	if err != nil {
		return &model.Order{}, err
	}
	return order, nil
}

func (s *OrderServiceImpl) Pull() error {
	orders, err := s.repos.GetAll()
	if err != nil {
		fmt.Println("aaaaaaa")
		return err
	}
	for _, order := range orders {
		s.cache.Put(order)
	}
	return nil
}
