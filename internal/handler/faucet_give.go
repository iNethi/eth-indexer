package handler

import (
	"context"
	"fmt"
	"math/big"

	"github.com/grassrootseconomics/eth-indexer/internal/telegram"
	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

const balanceThreshold = 50

func (h *Handler) IndexFaucetGive(ctx context.Context, event event.Event) error {
	return h.store.InsertFaucetGive(ctx, event)
}

func (h *Handler) FaucetHealthCheck(ctx context.Context, event event.Event) error {
	var balance *big.Int

	if err := h.chainProvider.Client.CallCtx(
		ctx,
		eth.Balance(w3.A(event.ContractAddress), nil).Returns(&balance),
	); err != nil {
		return err
	}

	if balance.Cmp(new(big.Int).Mul(w3.BigEther, big.NewInt(balanceThreshold))) < 0 {
		return h.telegram.Notify(ctx, fmt.Sprintf("%s: %s", event.ContractAddress, telegram.NOTIFY_LOW_BALANCE_ON_GAS_FAUCET))
	}

	return nil
}
