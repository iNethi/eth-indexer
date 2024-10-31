package handler

import (
	"context"

	"github.com/grassrootseconomics/eth-tracker/pkg/event"
)

func (h *Handler) IndexTransfer(ctx context.Context, event event.Event) error {
	return h.store.InsertTokenTransfer(ctx, event)
}
