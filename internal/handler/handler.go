package handler

import (
	"log/slog"

	"github.com/grassrootseconomics/eth-indexer/internal/cache"
	"github.com/grassrootseconomics/eth-indexer/internal/store"
	"github.com/grassrootseconomics/eth-indexer/internal/telegram"
	"github.com/grassrootseconomics/ethutils"
)

type (
	HandlerOpts struct {
		Store         store.Store
		Cache         *cache.Cache
		ChainProvider *ethutils.Provider
		Telegram      *telegram.Telegram
		Logg          *slog.Logger
	}

	Handler struct {
		store         store.Store
		cache         *cache.Cache
		chainProvider *ethutils.Provider
		telegram      *telegram.Telegram
		logg          *slog.Logger
	}
)

func NewHandler(o HandlerOpts) *Handler {
	return &Handler{
		store:         o.Store,
		cache:         o.Cache,
		chainProvider: o.ChainProvider,
		telegram:      o.Telegram,
		logg:          o.Logg,
	}
}
