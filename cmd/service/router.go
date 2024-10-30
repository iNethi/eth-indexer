package main

import (
	"github.com/grassrootseconomics/eth-indexer/internal/handler"
	"github.com/grassrootseconomics/eth-indexer/pkg/router"
)

func bootstrapRouter(handlerContainer *handler.Handler) *router.Router {
	router := router.New()

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
		handlerContainer.IndexTokenMint,
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

	return router
}
