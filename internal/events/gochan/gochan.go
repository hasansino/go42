package gochan

import (
	"context"
	"log/slog"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

type GoChan struct {
	logger  *slog.Logger
	channel *gochannel.GoChannel
	subwg   sync.WaitGroup
}

func New(opts ...Option) *GoChan {
	var (
		engine = new(GoChan)
	)

	for _, opt := range opts {
		opt(engine)
	}

	if engine.logger == nil {
		engine.logger = slog.New(slog.DiscardHandler)
	}

	goch := gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: 999,
			Persistent:          true,
		},
		watermill.NewSlogLogger(engine.logger),
	)

	return &GoChan{channel: goch}
}

func (g *GoChan) Publish(topic string, event []byte) error {
	msg := message.NewMessage(watermill.NewUUID(), event)
	return g.channel.Publish(topic, msg)
}

func (g *GoChan) Subscribe(
	ctx context.Context, topic string,
	handler func(ctx context.Context, event []byte) error,
) error {
	messages, err := g.channel.Subscribe(ctx, topic)
	if err != nil {
		return err
	}
	g.subwg.Add(1)
	go func() {
		for msg := range messages {
			err := handler(ctx, msg.Payload)
			if err != nil {
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
		g.subwg.Done()
	}()
	return nil
}

func (g *GoChan) Shutdown(_ context.Context) error {
	err := g.channel.Close()
	g.subwg.Wait()
	return err
}
