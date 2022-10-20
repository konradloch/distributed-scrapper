package broker

import (
	"context"
	"fmt"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yamlv3"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"time"
)

type configRmqList struct {
	Uri          string `mapstructure:"uri"`
	ExchangeName string `mapstructure:"exchangeName"`
	ExchangeType string `mapstructure:"exchangeType"`
	QueueName    string `mapstructure:"queueName"`
	Tag          string `mapstructure:"tag"`
}

type RabbitMQ struct {
	logger   *zap.SugaredLogger
	config   configRmqList
	conn     *amqp.Connection
	Delivery <-chan amqp.Delivery
}

func NewRabbitMq(logger *zap.SugaredLogger) *RabbitMQ {
	config.AddDriver(yamlv3.Driver)
	err := config.LoadFiles("pkg/config/config.yaml")
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
	d := initRabbitMqObjects(logger, connection, configmq)

	return &RabbitMQ{
		logger:   logger,
		config:   configmq,
		conn:     connection,
		Delivery: d,
	}
}

func initRabbitMqObjects(logger *zap.SugaredLogger, connection *amqp.Connection, configmq configRmqList) <-chan amqp.Delivery {
	logger.Infow("got Connection, getting Channel")
	channel, err := connection.Channel()
	if err != nil {
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
		"TEST",                // bindingKey
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
	return deliveries
}

func (r *RabbitMQ) Publish(routingKey, body string) error {
	channel, err := r.conn.Channel() //TODO verify how creating new channel affects performance
	if err != nil {
		panic(err)
	}
	defer channel.Close()
	var publishes chan uint64 = nil

	r.logger.Infow("enabling publisher confirms.")
	if err := channel.Confirm(false); err != nil {
		return fmt.Errorf("Channel could not be put into confirm mode: %s", err)
	}
	// We'll allow for a few outstanding publisher confirms
	publishes = make(chan uint64, 8)

	r.logger.Infow("declared Exchange, publishing messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	seqNo := channel.GetNextPublishSeqNo()
	r.logger.Infow("publishing %dB body (%q)", len(body), body)

	if err := channel.PublishWithContext(ctx,
		r.config.ExchangeName, // publish to an exchange
		routingKey,            // routing to 0 or more queues
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
		},
	); err != nil {
		return fmt.Errorf("Exchange Publish: %s", err)
	}

	r.logger.Infow("published %dB OK", len(body))
	publishes <- seqNo
	return nil
}
