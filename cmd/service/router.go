package main

import (
	"github.com/grassrootseconomics/eth-indexer/v2/internal/handler"
	"github.com/grassrootseconomics/eth-indexer/v2/pkg/router"
)

func bootstrapRouter(handlerContainer *handler.Handler) *router.Router {
	router := router.New(lo)

	router.RegisterRoute(
		"TRACKER.TOKEN_TRANSFER",
		handlerContainer.IndexTransfer,
		handlerContainer.AddToken,
	)
	router.RegisterRoute(
		"TRACKER.TOKEN_MINT",
		handlerContainer.IndexTokenMint,
		handlerContainer.AddToken,
	)
	router.RegisterRoute(
		"TRACKER.TOKEN_BURN",
		handlerContainer.IndexTokenBurn,
		handlerContainer.AddToken,
	)
	router.RegisterRoute(
		"TRACKER.POOL_SWAP",
		handlerContainer.IndexPoolSwap,
		handlerContainer.AddPool,
	)
	router.RegisterRoute(
		"TRACKER.POOL_DEPOSIT",
		handlerContainer.IndexPoolDeposit,
		handlerContainer.AddPool,
	)
	router.RegisterRoute(
		"TRACKER.FAUCET_GIVE",
		handlerContainer.IndexFaucetGive,
		handlerContainer.FaucetHealthCheck,
	)
	router.RegisterRoute(
		"TRACKER.OWNERSHIP_TRANSFERRED",
		handlerContainer.IndexOwnershipChange,
	)

	router.RegisterRoute(
		"TRACKER.INDEX_REMOVE",
		handlerContainer.IndexRemove,
	)

	return router
}
