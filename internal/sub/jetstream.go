package sub

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/grassrootseconomics/eth-indexer/pkg/router"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type (
	JetStreamOpts struct {
		Endpoint    string
		JetStreamID string
		Logg        *slog.Logger
		Router      *router.Router
	}

	JetStreamSub struct {
		jsConsumer jetstream.Consumer
		logg       *slog.Logger
		natsConn   *nats.Conn
		router     *router.Router
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
		router:     o.Router,
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
	iter, err := s.jsConsumer.Messages(jetstream.WithMessagesErrOnMissingHeartbeat(false))
	if err != nil {
		return err
	}
	defer iter.Stop()

	for {
		msg, err := iter.Next()
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				s.logg.Error("jetstream: iter fetch timeout")
				continue
			} else if errors.Is(err, nats.ErrConnectionClosed) {
				return nil
			} else {
				return err
			}
		}

		s.logg.Info("processing nats message", "subject", msg.Subject())
		if err := s.router.Handle(context.Background(), msg); err != nil {
			s.logg.Error("router: error processing nats message", "error", err)
		}
	}
}
