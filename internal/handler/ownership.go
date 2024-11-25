package handler

import (
	"context"

	"github.com/grassrootseconomics/eth-tracker/pkg/event"
)

func (h *Handler) IndexOwnershipChange(ctx context.Context, event event.Event) error {
	return h.store.InsertOwnershipChange(ctx, event)
}
