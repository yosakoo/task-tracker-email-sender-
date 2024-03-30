package rabbitmq

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yosakoo/task-tracker-email-sender-/internal/models"
	"github.com/yosakoo/task-tracker-email-sender-/internal/services"
)

type Config struct {
	URL          string
	Exchange     string
	ExchangeType string
	Queue        string
	BindingKey   string
	ConsumerTag  string
}


func New(cfg Config,lifetime *time.Duration,emailService *services.EmailService,) {
	c, err := NewConsumer(cfg.URL, cfg.Exchange, cfg.ExchangeType, cfg.Queue, cfg.BindingKey, cfg.ConsumerTag, emailService)
	if err != nil {
		log.Fatalf("%s", err)
	}

	SetupCloseHandler(c)

	if *lifetime > 0 {
		log.Printf("running for %s", *lifetime)
		time.Sleep(*lifetime)
	} else {
		log.Printf("running until Consumer is done")
		<-c.done
	}

	log.Printf("shutting down")

	if err := c.Shutdown(); err != nil {
		log.Fatalf("error during shutdown: %s", err)
	}
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
	emailService services.Emails
}

func SetupCloseHandler(consumer *Consumer) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Printf("rabbitmq stopped")
		if err := consumer.Shutdown(); err != nil {
			log.Fatalf("error during shutdown: %s", err)
		}
		os.Exit(0)
	}()
}

func NewConsumer(amqpURI, exchange, exchangeType, queueName, key, ctag string, emailService *services.EmailService) (*Consumer, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
		done:    make(chan error),
		emailService: emailService,
	}

	var err error

	config := amqp.Config{Properties: amqp.NewConnectionProperties()}
	config.Properties.SetClientConnectionName(c.tag)
	log.Print("dialing")
	c.conn, err = amqp.DialConfig(amqpURI, config)
	if err != nil {
		return nil, fmt.Errorf("dial: %s", err)
	}

	go func() {
		log.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Printf("got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("channel: %s", err)
	}

	log.Printf("got Channel, declaring Exchange (%q)", exchange)
	if err = c.channel.ExchangeDeclare(
		exchange,     // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return nil, fmt.Errorf("exchange Declare: %s", err)
	}

	log.Printf("declared Exchange, declaring Queue %q", queueName)
	queue, err := c.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("queue Declare: %s", err)
	}

	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, key)

	if err = c.channel.QueueBind(
		queue.Name, // name of the queue
		key,        // bindingKey
		exchange,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		return nil, fmt.Errorf("queue Bind: %s", err)
	}

	log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	deliveries, err := c.channel.Consume(
		queue.Name, // name
		c.tag,      // consumerTag,
		true,       // autoAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("queue Consume: %s", err)
	}

	go c.Handle(deliveries, c.done)

	return c, nil
}

func (c *Consumer) Shutdown() error {
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	return <-c.done
}

func (c *Consumer) Handle(deliveries <-chan amqp.Delivery, done chan error) {
	cleanup := func() {
		log.Printf("handle: deliveries channel closed")
		done <- nil
	}

	defer cleanup()

	for d := range deliveries {
        log.Printf(
            "got %dB delivery: [%v] %q",
            len(d.Body),
            d.DeliveryTag,
            d.Body)
        var msg models.EmailMessage
        err := json.Unmarshal(d.Body, &msg)
        if err != nil {
            log.Printf("error decoding JSON: %v", err)
            continue
        }
        err = c.emailService.SendEmail(msg.To, msg.Subject, msg.Body)
        if err != nil {
            log.Printf("error sending email: %v", err)
            continue
        }
    }
}
