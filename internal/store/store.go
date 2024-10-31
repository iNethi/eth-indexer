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
		InsertToken(context.Context, string, string, string, uint8, string) error
		InsertPool(context.Context, string, string, string) error
		Pool() *pgxpool.Pool
		Close()
	}
)
