package handler

import (
	"log/slog"

	"github.com/grassrootseconomics/eth-indexer/v2/internal/cache"
	"github.com/grassrootseconomics/eth-indexer/v2/internal/store"
	"github.com/grassrootseconomics/ethutils"
)

type (
	HandlerOpts struct {
		Store         store.Store
		Cache         *cache.Cache
		ChainProvider *ethutils.Provider
		Logg          *slog.Logger
	}

	Handler struct {
		store         store.Store
		cache         *cache.Cache
		chainProvider *ethutils.Provider
		logg          *slog.Logger
	}
)

func NewHandler(o HandlerOpts) *Handler {
	return &Handler{
		store:         o.Store,
		cache:         o.Cache,
		chainProvider: o.ChainProvider,
		logg:          o.Logg,
	}
}
