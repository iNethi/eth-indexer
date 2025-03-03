package handler

import (
	"log/slog"

	"github.com/grassrootseconomics/eth-indexer/v2/internal/cache"
	"github.com/grassrootseconomics/eth-indexer/v2/internal/store"
	"github.com/grassrootseconomics/eth-indexer/v2/pkg/inethi"
	"github.com/grassrootseconomics/ethutils"
)

type (
	HandlerOpts struct {
		VaultAddress  string
		Store         store.Store
		Cache         *cache.Cache
		ChainProvider *ethutils.Provider
		InethiClient  *inethi.InethiClient
		Logg          *slog.Logger
	}

	Handler struct {
		vaultAddress  string
		store         store.Store
		cache         *cache.Cache
		iClient       *inethi.InethiClient
		chainProvider *ethutils.Provider
		logg          *slog.Logger
	}
)

func NewHandler(o HandlerOpts) *Handler {
	return &Handler{
		vaultAddress:  o.VaultAddress,
		store:         o.Store,
		cache:         o.Cache,
		iClient:       o.InethiClient,
		chainProvider: o.ChainProvider,
		logg:          o.Logg,
	}
}
