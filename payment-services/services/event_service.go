package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"payment-services/models"
	"payment-services/repositories"

	"github.com/IBM/sarama"
)

type TransactionEvent struct {
	RequestID   string    `json:"request_id"`
	MerchantID  string    `json:"merchant_id"`
	BillNumber  string    `json:"bill_number"`
	Status      string    `json:"status"`
	PublishedAt time.Time `json:"published_at"`
}

// producer & consumer Kafka
type EventService struct {
	producer           sarama.AsyncProducer
	topic              string
	logger             *slog.Logger
	eventLogRepository *repositories.TransactionEventLogRepository
}

var ErrEventQueueFull = errors.New("event queue is full")

func NewEventService(
	producer sarama.AsyncProducer,
	topic string,
	logger *slog.Logger,
	eventLogRepository *repositories.TransactionEventLogRepository,
) *EventService {
	service := &EventService{
		producer:           producer,
		topic:              topic,
		logger:             logger,
		eventLogRepository: eventLogRepository,
	}
	go service.startProducerMonitor()
	return service
}

// worker
func (service *EventService) startProducerMonitor() {
	for {
		select {
		case msg, ok := <-service.producer.Successes():
			if !ok {
				return
			}
			if msg != nil {
				service.logger.Debug("transaction event published", "topic", msg.Topic)
			}
		case errMsg, ok := <-service.producer.Errors():
			if !ok {
				return
			}
			if errMsg != nil {
				service.logger.Error("failed to publish transaction event", "topic", errMsg.Msg.Topic, "error", errMsg.Err)
			}
		}
	}
}

func (service *EventService) PublishTransaction(ctx context.Context, event TransactionEvent) error {
	// convert event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	message := &sarama.ProducerMessage{
		Topic: service.topic,
		Key:   sarama.StringEncoder(event.RequestID),
		Value: sarama.ByteEncoder(payload),
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case service.producer.Input() <- message:
		return nil
	default:
		service.logger.Warn("event dropped due to full async producer queue", "topic", service.topic, "request_id", event.RequestID)
		return ErrEventQueueFull
	}
}

// consumer read event and save to database for async audit trail
func (service *EventService) StartTransactionConsumer(ctx context.Context, brokers []string) error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V3_7_0_0

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return err
	}

	partitions, err := consumer.Partitions(service.topic)
	if err != nil {
		return err
	}

	for _, partition := range partitions {
		partitionConsumer, err := consumer.ConsumePartition(service.topic, partition, sarama.OffsetNewest)
		if err != nil {
			return err
		}

		go func(pc sarama.PartitionConsumer) {
			defer pc.Close()
			for {
				select {
				case <-ctx.Done():
					return
				case message := <-pc.Messages():
					if message != nil {
						service.processEventMessage(ctx, message)
					}
				case err := <-pc.Errors():
					if err != nil {
						service.logger.Error("transaction event consumer error", "error", err.Err)
					}
				}
			}
		}(partitionConsumer)
	}

	return nil
}

// processing event from kafka consumer
func (service *EventService) processEventMessage(ctx context.Context, message *sarama.ConsumerMessage) {
	var event TransactionEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		service.logger.Error("failed to unmarshal transaction event", "topic", service.topic, "error", err)
		return
	}

	service.logger.Info("transaction event consumed from kafka", "request_id", event.RequestID, "merchant_id", event.MerchantID)

	// save event log to database for audit trail
	if service.eventLogRepository != nil {
		eventLog := &models.TransactionEventLog{
			RequestID:    event.RequestID,
			MerchantID:   event.MerchantID,
			BillNumber:   event.BillNumber,
			Status:       event.Status,
			EventPayload: models.JSONB(message.Value),
			PublishedAt:  event.PublishedAt,
			ConsumedAt:   time.Now().UTC(),
		}

		if err := service.eventLogRepository.CreateEventLog(ctx, eventLog); err != nil {
			service.logger.Error("failed to save transaction event log to database", "request_id", event.RequestID, "error", err)
			return
		}

		service.logger.Debug("transaction event log saved to database for audit trail", "request_id", event.RequestID, "bill_number", event.BillNumber)
	}
}
