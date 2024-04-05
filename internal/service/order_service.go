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
	// реализация создания заказа
	newOrder, err := s.repos.Create(order)
	orderFromDB, err1 := s.repos.GetOne(newOrder.OrderUUID)
	if err1 != nil {
		fmt.Errorf("error: %v", err1)
	}
	if err != nil {
		return model.Order{}, err
	}
	s.cache.Put(orderFromDB)
	return *newOrder, nil
}

func (s *OrderServiceImpl) GetOrder(uuid uuid.UUID) (*model.Order, error) {
	// реализация получения информации о заказе
	order, err := s.cache.Get(uuid)
	if err != nil {
		// при падении сервиса подтягиваем данные из бд в кеш
		orderFromDB, err1 := s.repos.GetOne(uuid)
		if err1 != nil {
			return &model.Order{}, err
		}
		s.cache.Put(orderFromDB)
		newOrderFromCache, err2 := s.cache.Get(orderFromDB.OrderUUID)
		if err2 != nil {
			return &model.Order{}, err2
		}
		return newOrderFromCache, nil
	}
	return order, nil
}

func (s *OrderServiceImpl) Pull() error {
	orders, err := s.repos.GetAll()
	if err != nil {
		return err
	}
	for _, order := range orders {
		s.cache.Put(&order)
	}
	return nil
}
