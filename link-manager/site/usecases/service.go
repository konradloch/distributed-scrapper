package usecases

import (
	"encoding/json"
	"fmt"
	"github.com/konradloch/distributed-scrapper/link-manager/site/broker"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Service struct {
	broker *broker.RabbitMQ
	logger *zap.SugaredLogger
}

func NewService(broker *broker.RabbitMQ, logger *zap.SugaredLogger) *Service {
	return &Service{
		broker: broker,
		logger: logger,
	}
}

func (s *Service) StartListening() {
	s.logger.Infow("start listening")
	for d := range s.broker.Delivery {
		go func(d *amqp091.Delivery) {
			var e broker.SiteResponseEvent
			if err := json.Unmarshal(d.Body, &e); err != nil {
				fmt.Println("problem with unmarshalling event")
				err := d.Reject(false)
				if err != nil {
					fmt.Println("problem with nack")
				}
				return
			}
			s.logger.Infow("response", "body", e)
		}(&d)
	}
}

func (s *Service) Publish(url string) error {
	return s.broker.Publish("TEST", broker.SiteRequestEvent{
		Url: url,
	})
}
