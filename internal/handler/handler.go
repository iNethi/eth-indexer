package handler

import (
	"context"
	"encoding/json"

	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/eth-indexer/internal/store"
)

type (
	HandlerOpts struct {
		Store store.Store
	}

	Handler struct {
		store store.Store
	}
)

func NewHandler(o HandlerOpts) *Handler {
	return &Handler{
		store: o.Store,
	}
}

func (h *Handler) Handle(ctx context.Context, msgSubject string, msgData []byte) error {
	var chainEvent event.Event

	if err := json.Unmarshal(msgData, &chainEvent); err != nil {
		return err
	}

	switch msgSubject {
	case "TRACKER.TOKEN_TRANSFER":
		return h.store.InsertTokenTransfer(ctx, chainEvent)
	case "TRACKER.POOL_SWAP":
		return h.store.InsertPoolSwap(ctx, chainEvent)
	case "TRACKER.FAUCET_GIVE":
		return h.store.InsertFaucetGive(ctx, chainEvent)
	case "TRACKER.POOL_DEPOSIT":
		return h.store.InsertPoolDeposit(ctx, chainEvent)
	case "TRACKER.TOKEN_MINT":
		return h.store.InsertTokenMint(ctx, chainEvent)
	case "TRACKER.TOKEN_BURN":
		return h.store.InsertTokenBurn(ctx, chainEvent)
	case "TRACKER.QUOTER_PRICE_INDEX_UPDATED":
		return h.store.InsertPriceQuoteUpdate(ctx, chainEvent)
	}

	return nil
}
