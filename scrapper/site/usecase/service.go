package usecase

import (
	"encoding/json"
	"fmt"
	"github.com/konradloch/distributed-scrapper/scrapper/site/broker"
	"github.com/konradloch/distributed-scrapper/scrapper/site/client"
	"github.com/rabbitmq/amqp091-go"
)

type Service struct {
	broker *broker.RabbitMQ
	wiki   *client.Wiki
}

func NewService(broker *broker.RabbitMQ, wiki *client.Wiki) *Service {
	return &Service{
		broker: broker,
		wiki:   wiki,
	}
}

func (s *Service) StartListening() {
	fmt.Println("start listening")
	for d := range s.broker.Delivery {
		go func(d *amqp091.Delivery) {
			var e broker.SiteRequestEvent
			if err := json.Unmarshal(d.Body, &e); err != nil {
				fmt.Println("problem with unmarshalling event")
				err := d.Reject(false)
				if err != nil {
					fmt.Println("problem with nack")
				}
				return
			}

			//TODO handle when root
			//cate, err := s.wiki.GetLinksRoot(e.Url)
			//if err != nil {
			//	fmt.Println("message will be nack due to wiki problem")
			//	err := d.Nack(false, true)
			//	if err != nil {
			//		fmt.Println("problem with nack")
			//	}
			//	return
			//}

			cate, arti, err := s.wiki.GetLinks(e.Url)
			if err != nil {
				fmt.Println("message will be nack due to wiki problem")
				err := d.Nack(false, true)
				if err != nil {
					fmt.Println("problem with nack")
				}
				return
			}
			resp := broker.SiteResponseEvent{
				ParentUrl:  e.Url,
				Categories: cate,
				Articles:   arti,
			}
			err = s.broker.Publish("HANDLE", resp)
			if err != nil {
				fmt.Println("message will be nack due to publish problem")
				err := d.Nack(false, true)
				if err != nil {
					fmt.Println("problem with nack")
				}
				return
			}
			err = d.Ack(false)
			if err != nil {
				fmt.Println("problem with ack")
				return
			}
			return
		}(&d)
	}
}
