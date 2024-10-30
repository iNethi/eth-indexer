package router

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sourcegraph/conc/pool"
)

type (
	HandlerFunc func(context.Context, event.Event) error

	Router struct {
		logg     *slog.Logger
		handlers map[string][]HandlerFunc
	}
)

func New() *Router {
	return &Router{
		handlers: make(map[string][]HandlerFunc),
	}
}

func (r *Router) RegisterRoute(subject string, handlerFunc ...HandlerFunc) {
	r.handlers[subject] = handlerFunc
}

func (r *Router) Handle(ctx context.Context, msg jetstream.Msg) error {
	handlers, ok := r.handlers[msg.Subject()]
	if !ok {
		return nil
	}

	var chainEvent event.Event
	if err := json.Unmarshal(msg.Data(), &chainEvent); err != nil {
		return err
	}

	p := pool.New().WithErrors()

	for _, handler := range handlers {
		p.Go(func() error {
			return handler(ctx, chainEvent)
		})
	}

	if err := p.Wait(); err != nil {
		return msg.Nak()
	}

	return msg.Ack()
}
