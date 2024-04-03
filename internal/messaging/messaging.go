package messaging

import (
	"encoding/json"
	"log"

	"github.com/mainpluv/wb_l0/internal/model"
	"github.com/mainpluv/wb_l0/internal/service"
	"github.com/nats-io/stan.go"
)

type Subscriber struct {
	sc      stan.Conn
	service service.OrderService
}
type Publisher struct {
	sc stan.Conn
}

func NewSubscriber(sc stan.Conn, service service.OrderService) *Subscriber {
	return &Subscriber{sc: sc, service: service}
}
func NewPublisher(sc stan.Conn) *Publisher {
	return &Publisher{sc: sc}
}

func (s *Subscriber) StartSubscriber() error {
	_, err := s.sc.Subscribe("try1", func(msg *stan.Msg) {
		if err := s.parseMessage(msg.Data); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	})
	if err != nil {
		log.Printf("Error subscribing to channel: %v", err)
		return err
	}
	return nil

}

func (p *Publisher) StartPublisher() error {
	for i := 0; i < 10; i++ {
		err := p.publishMessege()
		if err != nil {
			log.Printf("Error publishing message: %v", err)
			return err
		}
	}
	return nil
}

func (p *Publisher) publishMessege() error {
	order := GenerateOrder()
	orderJSON, err := json.Marshal(order)
	if err != nil {
		log.Printf("Error marshaling order to JSON: %v", err)
		return err
	}
	err = p.sc.Publish("try1", orderJSON)
	if err != nil {
		log.Printf("Error publishing message: %v", err)
		return err
	}
	return nil

}

func (s *Subscriber) parseMessage(data []byte) error {
	order := &model.Order{}
	err := json.Unmarshal(data, order)
	if err != nil {
		log.Printf("Error parsing message: %v", err)
		return err
	}
	_, err = s.service.SaveOrder(*order)
	if err != nil {
		log.Printf("Error saving order: %v", err)
		return err
	}
	return nil
}
