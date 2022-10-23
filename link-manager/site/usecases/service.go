package usecases

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/konradloch/distributed-scrapper/link-manager/site/broker"
	"github.com/konradloch/distributed-scrapper/link-manager/site/repository"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Service struct {
	broker *broker.RabbitMQ
	logger *zap.SugaredLogger
	repo   *repository.SiteRepository
}

func NewService(broker *broker.RabbitMQ, repo *repository.SiteRepository, logger *zap.SugaredLogger) *Service {
	return &Service{
		broker: broker,
		logger: logger,
		repo:   repo,
	}
}

func (s *Service) StartListening() {
	s.logger.Infow("start listening")
	for d := range s.broker.Delivery {
		func(d *amqp091.Delivery) {
			isAck, retry := func() (bool, bool) {
				var e broker.SiteResponseEvent
				if err := json.Unmarshal(d.Body, &e); err != nil {
					s.logger.Errorw("problem with unmarshalling event")
					return false, false
				}
				s.logger.Infow("got response from scrapper")
				parent, err := s.repo.GetByUrl(e.ParentUrl)
				if err != nil {
					s.logger.Errorw("problem with fetching parent")
					return false, true
				}
				for _, c := range e.Categories {
					err := s.Publish(c, "CATEGORY", &parent.ID)
					if err != nil {
						s.logger.Errorw("cannot publish message")
						return false, true
					}
				}
				for _, a := range e.Articles {
					byUrl, err := s.repo.GetByUrl(a)
					if err != nil {
						s.logger.Errorw("article record fetching problem")
						return false, true
					}
					if byUrl != nil {
						s.logger.Errorw("article already exist in database", "url", a)
						return false, false
					}
					_, err = s.repo.Save(a, "ARTICLES", &parent.ID)
					if err != nil {
						s.logger.Errorw("cannot save articles")
						return false, true
					}
				}
				err = s.repo.UpdateStatusByUrl(parent.Url, "PROCESSED")
				if err != nil {
					s.logger.Error("fail to change status", "err", err)
					return false, true
				}
				return true, false
			}()
			if isAck {
				err := d.Ack(false)
				if err != nil {
					s.logger.Errorw("failed to ack", "err", err)
					return
				}
			} else {
				err := d.Nack(false, retry)
				if err != nil {
					s.logger.Errorw("failed to nack", "err", err)
					return
				}
			}
		}(&d)
	}
}

func (s *Service) Publish(url, category string, parentID *uuid.UUID) error {
	byUrl, err := s.repo.GetByUrl(url)
	if err != nil {
		s.logger.Infow("record not found")
		return err
	}
	if byUrl != nil {
		s.logger.Infow("already exist in database", "url", url)
		return nil
	}
	_, err = s.repo.Save(url, category, parentID)
	if err != nil {
		return err
	}
	return s.broker.Publish("TEST", broker.SiteRequestEvent{
		Url: url,
	})
}
