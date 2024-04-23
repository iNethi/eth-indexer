package store

import (
	"context"

	"github.com/grassrootseconomics/celo-indexer/internal/event"
)

type (
	Store interface {
		InsertTokenTransfer(context.Context, event.Event) error
		InsertTokenMint(context.Context, event.Event) error
		InsertPoolSwap(context.Context, event.Event) error
		InsertPoolDeposit(context.Context, event.Event) error
	}
)
