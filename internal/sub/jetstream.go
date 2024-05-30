package sub

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/grassrootseconomics/celo-indexer/internal/store"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type (
	JetStreamOpts struct {
		Store       store.Store
		Logg        *slog.Logger
		Endpoint    string
		JetStreamID string
	}

	JetStreamSub struct {
		jsConsumer jetstream.Consumer
		store      store.Store
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
		Durable:       o.JetStreamID,
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: pullStream,
	})
	if err != nil {
		return nil, err
	}
	o.Logg.Info("successfully connected to NATS server")

	return &JetStreamSub{
		jsConsumer: consumer,
		store:      o.Store,
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
			if err := s.processEventHandler(context.Background(), msg.Subject(), msg.Data()); err != nil {
				s.logg.Error("error processing nats message", "error", err)
				msg.Nak()
			} else {
				msg.Ack()
			}
		}
	}
}

func (s *JetStreamSub) processEventHandler(ctx context.Context, msgSubject string, msgData []byte) error {
	var chainEvent event.Event

	if err := json.Unmarshal(msgData, &chainEvent); err != nil {
		return err
	}

	switch msgSubject {
	case "TRACKER.TOKEN_TRANSFER":
		if err := s.store.InsertTokenTransfer(ctx, chainEvent); err != nil {
			return err
		}
	case "TRACKER.TOKEN_MINT":
		if err := s.store.InsertTokenMint(ctx, chainEvent); err != nil {
			return err
		}
	case "TRACKER.POOL_SWAP":
		if err := s.store.InsertPoolSwap(ctx, chainEvent); err != nil {
			return err
		}
	case "TRACKER.POOL_DEPOSIT":
		if err := s.store.InsertPoolDeposit(ctx, chainEvent); err != nil {
			return err
		}
	}

	return nil
}
