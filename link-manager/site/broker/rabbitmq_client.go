package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yamlv3"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"time"
)

type configRmqList struct {
	Uri                  string `mapstructure:"uri"`
	ExchangeName         string `mapstructure:"exchangeName"`
	ExchangeType         string `mapstructure:"exchangeType"`
	QueueName            string `mapstructure:"queueName"`
	Tag                  string `mapstructure:"tag"`
	ProducerExchangeName string `mapstructure:"producerExchangeName"`
}

type RabbitMQ struct {
	logger   *zap.SugaredLogger
	config   configRmqList
	conn     *amqp.Connection
	Delivery <-chan amqp.Delivery
	Channel  *amqp.Channel
}

func NewRabbitMQ(logger *zap.SugaredLogger) *RabbitMQ {
	config.WithOptions(config.ParseEnv)
	config.AddDriver(yamlv3.Driver)
	err := config.LoadFiles("config/config.yaml")
	if err != nil {
		panic(err)
	}
	configmq := configRmqList{}
	err = config.BindStruct("rabbitmq", &configmq)
	if err != nil {
		panic(err)
	}

	c := amqp.Config{Properties: amqp.NewConnectionProperties()}
	c.Properties.SetClientConnectionName("sample-producer")
	logger.Infow("dialing", "uri", configmq.Uri)
	connection, err := amqp.DialConfig(configmq.Uri, c)
	if err != nil {
		panic(err)
	}
	d, ch := initRabbitMqObjects(logger, connection, configmq)

	return &RabbitMQ{
		logger:   logger,
		config:   configmq,
		conn:     connection,
		Delivery: d,
		Channel:  ch,
	}
}

func initRabbitMqObjects(logger *zap.SugaredLogger, connection *amqp.Connection, configmq configRmqList) (<-chan amqp.Delivery, *amqp.Channel) {
	logger.Infow("got Connection, getting Channel")
	channel, err := connection.Channel()
	if err != nil {
		panic(err)
	}

	logger.Infow("got Channel, declaring %q Exchange (%q)", "exchangeType", configmq.ExchangeType, "exchangeName", configmq.ExchangeName)
	if err := channel.ExchangeDeclare(
		configmq.ProducerExchangeName, // name
		configmq.ExchangeType,         // type
		true,                          // durable
		false,                         // auto-deleted
		false,                         // internal
		false,                         // noWait
		nil,                           // arguments
	); err != nil {
		panic(err)
	}

	logger.Infow("got Channel, declaring %q Exchange (%q)", "exchangeType", configmq.ExchangeType, "exchangeName", configmq.ExchangeName)
	if err := channel.ExchangeDeclare(
		configmq.ExchangeName, // name
		configmq.ExchangeType, // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // noWait
		nil,                   // arguments
	); err != nil {
		panic(err)
	}

	queue, err := channel.QueueDeclare(
		configmq.QueueName, // name of the queue
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // noWait
		nil,                // arguments
	)
	if err != nil {
		panic(err)
	}

	if err = channel.QueueBind(
		queue.Name,            // name of the queue
		"HANDLE",              // bindingKey
		configmq.ExchangeName, // sourceExchange
		false,                 // noWait
		nil,                   // arguments
	); err != nil {
		panic(err)
	}

	deliveries, err := channel.Consume(
		queue.Name,   // name
		configmq.Tag, // consumerTag,
		false,        // autoAck
		false,        // exclusive
		false,        // noLocal
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		panic(err)
	}
	return deliveries, channel
}

// TODO consider use another exachange, will it affect performance? - DONE but check performance
func (r *RabbitMQ) Publish(routingKey string, msg interface{}) error {
	r.logger.Infow("declared Exchange, publishing messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if err := r.Channel.PublishWithContext(ctx,
		r.config.ProducerExchangeName, // publish to an exchange
		routingKey,                    // routing to 0 or more queues
		false,                         // mandatory
		false,                         // immediate
		amqp.Publishing{
			Headers:      amqp.Table{},
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // 1=non-persistent, 2=persistent
			//Priority:     0,               // 0-9
		},
	); err != nil {
		return fmt.Errorf("exchange Publish: %s", err)
	}

	r.logger.Infow("published OK")
	return nil
}
