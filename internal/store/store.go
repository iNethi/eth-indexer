package store

import (
	"context"

	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Store interface {
		InsertTokenTransfer(context.Context, event.Event) error
		InsertTokenMint(context.Context, event.Event) error
		InsertTokenBurn(context.Context, event.Event) error
		InsertFaucetGive(context.Context, event.Event) error
		InsertPoolSwap(context.Context, event.Event) error
		InsertPoolDeposit(context.Context, event.Event) error
		InsertPriceQuoteUpdate(context.Context, event.Event) error
		Pool() *pgxpool.Pool
		Close()
	}
)
