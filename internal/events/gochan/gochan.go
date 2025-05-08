package gochan

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

type GoChan struct {
	engine *gochannel.GoChannel
}

func New() *GoChan {
	engine := gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: 999,
			Persistent:          true,
		},
		watermill.NewSlogLogger(
			slog.Default().With(slog.String("component", "events-gochan")),
		),
	)
	return &GoChan{engine: engine}
}

func (g *GoChan) Publish(topic string, event []byte) error {
	msg := message.NewMessage(watermill.NewUUID(), event)
	return g.engine.Publish(topic, msg)
}

func (g *GoChan) Subscribe(
	ctx context.Context, topic string,
	handler func(ctx context.Context, event []byte) error,
) error {
	messages, err := g.engine.Subscribe(ctx, topic)
	if err != nil {
		return err
	}
	go func() {
		for msg := range messages {
			err := handler(ctx, msg.Payload)
			if err != nil {
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}()
	return nil
}
