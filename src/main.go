package main

import (
	"fmt"
	"log"

	"github.com/nats-io/stan.go"

	"github.com/mainpluv/wb_l0/internal/cache"
	"github.com/mainpluv/wb_l0/internal/database"
	"github.com/mainpluv/wb_l0/internal/delivery"
	"github.com/mainpluv/wb_l0/internal/messaging"
	"github.com/mainpluv/wb_l0/internal/service"
)

func main() {
	connString := "postgres://l0_user:L0@localhost:5432/L0_db"
	pool, err := database.NewPool(connString)
	if err != nil {
		log.Fatalf("Error creating db pool: %v", err)
	}
	log.Println("successful connection to db")
	defer pool.Close()

	orderRepo := database.NewOrderRepo(pool)
	memCache := cache.NewMemoryCache()
	orderService := service.NewOrderService(orderRepo, memCache)
	if err := orderService.Pull(); err != nil {
		fmt.Printf("Failed to pull orders: %v", err.Error())
	}
	handler := delivery.NewHandler(orderService)
	router := handler.InitRoutes()
	server := delivery.NewServer(router)
	sc, err := stan.Connect("my_cluster", "client_id", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		log.Fatalf("Failed to connect to nats-streaming server: %v", err)
	}
	defer sc.Close()
	sub := messaging.NewSubscriber(sc, orderService)
	sub.StartSubscriber()
	pub := messaging.NewPublisher(sc)
	pub.StartPublisher()
	log.Println("Server started")
	if err := server.RunServer(); err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
}
