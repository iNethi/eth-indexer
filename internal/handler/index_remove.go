package handler

import (
	"context"

	"github.com/grassrootseconomics/eth-tracker/pkg/event"
)

func (h *Handler) IndexRemove(ctx context.Context, event event.Event) error {
	return h.store.RemoveContractAddress(ctx, event)
}
