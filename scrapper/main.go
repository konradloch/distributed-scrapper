package main

import (
	broker2 "github.com/konradloch/distributed-scrapper/scrapper/site/broker"
	"github.com/konradloch/distributed-scrapper/scrapper/site/client"
	"github.com/konradloch/distributed-scrapper/scrapper/site/usecase"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()
	wiki := client.NewWiki(sugar)
	broker := broker2.NewRabbitMQ(sugar)
	service := usecase.NewService(broker, wiki)
	service.StartListening()
}
