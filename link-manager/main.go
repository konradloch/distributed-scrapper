package main

import (
	"github.com/konradloch/distributed-scrapper/link-manager/site/api"
	"github.com/konradloch/distributed-scrapper/link-manager/site/broker"
	"github.com/konradloch/distributed-scrapper/link-manager/site/usecases"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()
	brokerClient := broker.NewRabbitMQ(sugar)
	service := usecases.NewService(brokerClient, sugar)
	s := api.NewHttpServer(service)
	go service.StartListening()
	s.StartServer()
}
