package handler

import (
	"context"

	"github.com/grassrootseconomics/eth-tracker/pkg/event"
)

func (h *Handler) IndexPoolDeposit(ctx context.Context, event event.Event) error {
	return h.store.InsertPoolDeposit(ctx, event)
}
