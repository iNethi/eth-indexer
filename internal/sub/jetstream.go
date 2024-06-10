package sub

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/grassrootseconomics/celo-indexer/internal/handler"
	"github.com/grassrootseconomics/celo-indexer/internal/store"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type (
	JetStreamOpts struct {
		Store       store.Store
		Logg        *slog.Logger
		Handler     *handler.Handler
		Endpoint    string
		JetStreamID string
	}

	JetStreamSub struct {
		jsConsumer jetstream.Consumer
		store      store.Store
		handler    *handler.Handler
		natsConn   *nats.Conn
		logg       *slog.Logger
		durableID  string
	}
)

const (
	pullStream  = "TRACKER"
	pullSubject = "TRACKER.*"
)

func NewJetStreamSub(o JetStreamOpts) (Sub, error) {
	natsConn, err := nats.Connect(o.Endpoint)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(natsConn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := js.Stream(ctx, pullStream)
	if err != nil {
		return nil, err
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:   o.JetStreamID,
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return nil, err
	}
	o.Logg.Info("successfully connected to NATS server")

	return &JetStreamSub{
		jsConsumer: consumer,
		store:      o.Store,
		handler:    o.Handler,
		natsConn:   natsConn,
		logg:       o.Logg,
		durableID:  o.JetStreamID,
	}, nil
}

func (s *JetStreamSub) Close() {
	if s.natsConn != nil {
		s.natsConn.Close()
	}
}

func (s *JetStreamSub) Process() error {
	for {
		events, err := s.jsConsumer.Fetch(100, jetstream.FetchMaxWait(1*time.Second))
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				continue
			} else if errors.Is(err, nats.ErrConnectionClosed) {
				return nil
			} else {
				return err
			}
		}

		for msg := range events.Messages() {
			if err := s.handler.Handle(context.Background(), msg.Subject(), msg.Data()); err != nil {
				s.logg.Error("error processing nats message", "error", err)
				msg.Nak()
			} else {
				msg.Ack()
			}
		}
	}
}
