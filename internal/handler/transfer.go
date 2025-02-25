package handler

import (
	"context"
	"fmt"
	"math/big"

	"github.com/grassrootseconomics/eth-indexer/v2/pkg/inethi"
	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

func (h *Handler) IndexTransfer(ctx context.Context, event event.Event) error {
	return h.store.InsertTokenTransfer(ctx, event)
}

const (
	SIZE_1_GB   = 1
	SIZE_10_GB  = 2
	SIZE_40_GB  = 4
	SIZE_100_GB = 3
)

var (
	// 5 DUNIA
	PRICE_1_GB = big.NewInt(5_000_000)
	// 40 DUNIA
	PRICE_10_GB = big.NewInt(40_000_000)
	// 120 DUNIA
	PRICE_40_GB = big.NewInt(120_000_000)
	// 210 DUNIA
	PRICE_210_GB = big.NewInt(210_000_000)

	AMOUNT_DIVISOR = new(big.Float).SetInt(big.NewInt(1_000_000))
)

func (h *Handler) GenerateVoucher(ctx context.Context, event event.Event) error {
	recipientAddress := event.Payload["to"].(string)
	h.logg.Debug("generate voucher", "recipient", recipientAddress)
	if recipientAddress != h.vaultAddress {
		return nil
	}

	voucherPayload := inethi.VoucherPayload{
		SenderAddress:    event.Payload["from"].(string),
		RecipientAddress: recipientAddress,
	}

	rec, _ := new(big.Int).SetString(event.Payload["value"].(string), 10)
	h.logg.Debug("generate voucher", "amount", rec)

	if rec.Cmp(PRICE_1_GB) == 0 || (rec.Cmp(PRICE_1_GB) == 1 && rec.Cmp(PRICE_10_GB) == -1) {
		voucherPayload.CouponSize = SIZE_1_GB
	} else if rec.Cmp(PRICE_10_GB) == 0 || (rec.Cmp(PRICE_10_GB) == 1 && rec.Cmp(PRICE_40_GB) == -1) {
		voucherPayload.CouponSize = SIZE_10_GB
	} else if rec.Cmp(PRICE_40_GB) == 0 || (rec.Cmp(PRICE_40_GB) == 1 && rec.Cmp(PRICE_210_GB) == -1) {
		voucherPayload.CouponSize = SIZE_40_GB
	} else if rec.Cmp(PRICE_210_GB) == 0 || rec.Cmp(PRICE_210_GB) == 1 {
		voucherPayload.CouponSize = SIZE_100_GB
	}
	voucherPayload.Amount = formatAmount(rec)
	h.logg.Debug("generate voucher", "amount", voucherPayload.Amount, "size", voucherPayload.CouponSize)

	if h.cache.Get(event.ContractAddress) {
		tokenSymbol, err := h.store.GetTokenSymbol(ctx, event.ContractAddress)
		if err != nil {
			return err
		}
		voucherPayload.TokenSymbol = tokenSymbol
	} else {
		var tokenSymbol string

		contractAddress := w3.A(event.ContractAddress)

		if err := h.chainProvider.Client.CallCtx(
			ctx,
			eth.CallFunc(contractAddress, symbolGetter).Returns(&tokenSymbol),
		); err != nil {
			return err
		}
		voucherPayload.TokenSymbol = tokenSymbol
	}

	resp, err := h.iClient.GenerateVoucher(
		ctx,
		voucherPayload,
	)
	if err != nil {
		return err
	}
	h.logg.Debug("voucher generated", "voucher", resp.Voucher, "sender", voucherPayload.SenderAddress, "recipient", voucherPayload.RecipientAddress, "amount", voucherPayload.Amount, "token", voucherPayload.TokenSymbol)
	return nil
}

func formatAmount(dividend *big.Int) string {
	floatDividend := new(big.Float).SetInt(dividend)
	result := new(big.Float).Quo(floatDividend, AMOUNT_DIVISOR)
	return fmt.Sprintf("%.8f", result)
}
