package sub

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/grassrootseconomics/celo-indexer/internal/event"
	"github.com/grassrootseconomics/celo-indexer/internal/store"
	"github.com/nats-io/nats.go"
)

const (
	durableId   = "celo-indexer-6"
	pullStream  = "TRACKER"
	pullSubject = "TRACKER.*"
)

type (
	JetStreamOpts struct {
		Logg     *slog.Logger
		Endpoint string
		Store    store.Store
	}

	JetStreamSub struct {
		natsConn *nats.Conn
		jsCtx    nats.JetStreamContext
		store    store.Store
		logg     *slog.Logger
	}
)

func NewJetStreamSub(o JetStreamOpts) (Sub, error) {
	natsConn, err := nats.Connect(o.Endpoint)
	if err != nil {
		return nil, err
	}

	js, err := natsConn.JetStream()
	if err != nil {
		return nil, err
	}
	o.Logg.Info("successfully connected to NATS server")

	_, err = js.AddConsumer(pullStream, &nats.ConsumerConfig{
		Durable:       durableId,
		AckPolicy:     nats.AckExplicitPolicy,
		FilterSubject: pullSubject,
	})
	if err != nil {
		return nil, err
	}

	return &JetStreamSub{
		natsConn: natsConn,
		jsCtx:    js,
		store:    o.Store,
		logg:     o.Logg,
	}, nil
}

func (s *JetStreamSub) Close() {
	if s.natsConn != nil {
		s.natsConn.Close()
	}
}

func (s *JetStreamSub) Process() error {
	subOpts := []nats.SubOpt{
		nats.ManualAck(),
		nats.Bind(pullStream, durableId),
	}

	natsSub, err := s.jsCtx.PullSubscribe(pullSubject, durableId, subOpts...)
	if err != nil {
		return err
	}

	for {
		events, err := natsSub.Fetch(1)
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				continue
			} else if errors.Is(err, nats.ErrConnectionClosed) {
				return nil
			} else {
				return err
			}
		}

		if len(events) > 0 {
			msg := events[0]
			if err := s.processEventHandler(context.Background(), msg); err != nil {
				s.logg.Error("error processing nats message", "error", err)
				msg.Nak()
			} else {
				msg.Ack()
			}
		}
	}
}

func (s *JetStreamSub) processEventHandler(ctx context.Context, msg *nats.Msg) error {
	var (
		chainEvent event.Event
	)

	if err := json.Unmarshal(msg.Data, &chainEvent); err != nil {
		return err
	}

	switch msg.Subject {
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
