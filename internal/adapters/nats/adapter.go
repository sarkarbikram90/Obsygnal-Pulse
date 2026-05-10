// Package nats implements the event streaming messaging layer using NATS JetStream.
//
// This adapter provides durable at-least-once delivery, consumer groups for
// horizontal scalability, and DLQ routing for poison messages.
package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
)

const (
	// SubjectGPU is the subject for GPU telemetry metrics
	SubjectGPU = "telemetry.gpu.v1"
	// SubjectNode is the subject for Node telemetry metrics
	SubjectNode = "telemetry.node.v1"
	// SubjectDLQ is the subject for Dead Letter Queue
	SubjectDLQ = "telemetry.dlq.v1"

	// ConsumerGroupIngester is the durable consumer group for workers
	ConsumerGroupIngester = "telemetry_ingester_group"
)

// Adapter implements the NATS JetStream messaging port.
type Adapter struct {
	nc  *nats.Conn
	js  nats.JetStreamContext
	log *logger.Logger
	cfg config.NatsConfig
}

// NewAdapter creates and initializes a new NATS JetStream adapter.
func NewAdapter(cfg config.NatsConfig, log *logger.Logger) (*Adapter, error) {
	// Connect to NATS
	opts := []nats.Option{
		nats.Name("obsygnal-pulse"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2 * time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Warn().Err(err).Msg("disconnected from NATS")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info().Msg("reconnected to NATS")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Warn().Msg("NATS connection closed")
		}),
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	adapter := &Adapter{
		nc:  nc,
		js:  js,
		log: log.WithComponent("nats"),
		cfg: cfg,
	}

	if err := adapter.setupStreams(); err != nil {
		nc.Close()
		return nil, err
	}

	return adapter, nil
}

// setupStreams ensures the required JetStream streams and consumers exist.
func (a *Adapter) setupStreams() error {
	streamName := a.cfg.StreamName

	// Create or update the main telemetry stream
	stream, err := a.js.StreamInfo(streamName)
	if err != nil {
		if err == nats.ErrStreamNotFound {
			_, err = a.js.AddStream(&nats.StreamConfig{
				Name:      streamName,
				Subjects:  []string{"telemetry.>"},
				Retention: nats.LimitsPolicy, // Keep messages until size/age limits are hit
				MaxAge:    7 * 24 * time.Hour,
				MaxBytes:  10 * 1024 * 1024 * 1024, // 10 GB
				Storage:   nats.FileStorage,
				Discard:   nats.DiscardOld,
				NoAck:     false,
			})
			if err != nil {
				return fmt.Errorf("failed to create stream: %w", err)
			}
			a.log.Info().Str("stream", streamName).Msg("created JetStream stream")
		} else {
			return fmt.Errorf("failed to get stream info: %w", err)
		}
	} else {
		a.log.Info().Str("stream", stream.Config.Name).Msg("JetStream stream already exists")
	}

	// Create durable consumer for the worker ingester group
	_, err = a.js.AddConsumer(streamName, &nats.ConsumerConfig{
		Durable:       ConsumerGroupIngester,
		DeliverPolicy: nats.DeliverAllPolicy,
		AckPolicy:     nats.AckExplicitPolicy,
		AckWait:       a.cfg.AckWait,
		MaxDeliver:    a.cfg.MaxRetries + 1,
		FilterSubject: "telemetry.>",
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	return nil
}

// PublishBatch publishes a batch of metrics to JetStream.
func (a *Adapter) PublishBatch(ctx context.Context, batch *metrics.MetricBatch) error {
	data, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to marshal batch: %w", err)
	}

	msg := &nats.Msg{
		Subject: SubjectGPU,
		Data:    data,
		Header:  nats.Header{},
	}
	msg.Header.Set("Source", batch.Source)

	// Publish synchronously for MVP to ensure it lands in the stream
	_, err = a.js.PublishMsg(msg)
	if err != nil {
		return fmt.Errorf("failed to publish message to JetStream: %w", err)
	}

	return nil
}

// RouteToDLQ routes a poison message to the Dead Letter Queue.
func (a *Adapter) RouteToDLQ(msg *nats.Msg, errReason string) error {
	dlqMsg := &nats.Msg{
		Subject: SubjectDLQ,
		Data:    msg.Data,
		Header:  nats.Header{},
	}
	dlqMsg.Header.Set("Error-Reason", errReason)
	dlqMsg.Header.Set("Original-Subject", msg.Subject)

	_, err := a.js.PublishMsg(dlqMsg)
	return err
}

// SubscribeIngester starts a worker loop consuming from the stream.
func (a *Adapter) SubscribeIngester(ctx context.Context, handler func(context.Context, *metrics.MetricBatch) error) error {
	sub, err := a.js.PullSubscribe("telemetry.>", ConsumerGroupIngester)
	if err != nil {
		return fmt.Errorf("failed to subscribe to JetStream: %w", err)
	}

	a.log.Info().Str("consumer", ConsumerGroupIngester).Msg("started JetStream consumer")

	go func() {
		for {
			select {
			case <-ctx.Done():
				sub.Unsubscribe()
				return
			default:
				// Fetch up to 100 messages at a time, blocking up to 2 seconds
				msgs, err := sub.Fetch(100, nats.MaxWait(2*time.Second))
				if err != nil {
					if err == nats.ErrTimeout {
						continue // No messages, try again
					}
					if err == nats.ErrConnectionClosed || err == nats.ErrBadSubscription {
						return // Shutting down
					}
					a.log.Warn().Err(err).Msg("failed to fetch messages from JetStream")
					continue
				}

				for _, msg := range msgs {
					meta, _ := msg.Metadata()

					var batch metrics.MetricBatch
					if err := json.Unmarshal(msg.Data, &batch); err != nil {
						a.log.Error().Err(err).Msg("failed to unmarshal message, routing to DLQ")
						a.RouteToDLQ(msg, "json_unmarshal_failed")
						msg.Ack()
						continue
					}

					// Process batch
					if err := handler(ctx, &batch); err != nil {
						a.log.Error().Err(err).Msg("failed to process message")

						// If we've hit max retries, route to DLQ
						if int(meta.NumDelivered) >= a.cfg.MaxRetries {
							a.log.Warn().Str("subject", msg.Subject).Msg("max retries exceeded, routing to DLQ")
							a.RouteToDLQ(msg, fmt.Sprintf("max_retries_exceeded: %v", err))
							msg.Ack() // Ack original so it doesn't get redelivered again
						} else {
							msg.Nak() // Nak so it gets redelivered after AckWait
						}
						continue
					}

					// Successfully processed
					msg.Ack()
				}
			}
		}
	}()

	return nil
}

// Close closes the NATS connection.
func (a *Adapter) Close() error {
	a.nc.Close()
	return nil
}
