package handler

import (
	"context"
	"fmt"
	"math/big"

	"github.com/grassrootseconomics/eth-indexer/v2/pkg/inethi"
	"github.com/grassrootseconomics/eth-indexer/v2/pkg/notify"
	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

func (h *Handler) IndexTransfer(ctx context.Context, event event.Event) error {
	return h.store.InsertTokenTransfer(ctx, event)
}

const (
	SIZE_500_MB           = 25
	SIZE_1_GB             = 23
	SIZE_3_GB             = 24
	SIZE_5_GB             = 26
	SIZE_1_MONTH_HOME     = 35
	SIZE_1_MONTH_BUSINESS = 36
)

var (
	// 500 MB = 10 DUNIA
	PRICE_500_MB = big.NewInt(10_000_000)
	// 1 GB = 20 DUNIA
	PRICE_1_GB = big.NewInt(20_000_000)
	// 3 GB = 50 DUNIA
	PRICE_3_GB = big.NewInt(50_000_000)
	// 5 GB = 80 DUNIA
	PRICE_5_GB = big.NewInt(80_000_000)
	// 1 Month (Home) = 5000 DUNIA
	PRICE_1_MONTH_HOME = big.NewInt(2000_000_000)
	// 1 Month (Business) = 5000 DUNIA
	PRICE_1_MONTH_BUSINESS = big.NewInt(5000_000_000)

	AMOUNT_DIVISOR = new(big.Float).SetInt(big.NewInt(1_000_000))

	// Human-readable price tier descriptions
	PRICE_500_MB_DESC           = "500 MB"
	PRICE_1_GB_DESC             = "1 GB"
	PRICE_3_GB_DESC             = "3 GB"
	PRICE_5_GB_DESC             = "5 GB"
	PRICE_1_MONTH_HOME_DESC     = "1 Month Home Unlimited"
	PRICE_1_MONTH_BUSINESS_DESC = "1 Month Business Unlimited"
)

func (h *Handler) GenerateVoucher(ctx context.Context, event event.Event) error {
	if !event.Success {
		h.logg.Warn("tx reverted on chain", "tx_hash", event.TxHash)
		return nil
	}

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

	var tierDescription string
	switch {
	case rec.Cmp(PRICE_500_MB) == 0 || (rec.Cmp(PRICE_500_MB) == 1 && rec.Cmp(PRICE_1_GB) == -1):
		voucherPayload.CouponSize = SIZE_500_MB
		tierDescription = PRICE_500_MB_DESC
	case rec.Cmp(PRICE_1_GB) == 0 || (rec.Cmp(PRICE_1_GB) == 1 && rec.Cmp(PRICE_3_GB) == -1):
		voucherPayload.CouponSize = SIZE_1_GB
		tierDescription = PRICE_1_GB_DESC
	case rec.Cmp(PRICE_3_GB) == 0 || (rec.Cmp(PRICE_3_GB) == 1 && rec.Cmp(PRICE_5_GB) == -1):
		voucherPayload.CouponSize = SIZE_3_GB
		tierDescription = PRICE_3_GB_DESC
	case rec.Cmp(PRICE_5_GB) == 0 || (rec.Cmp(PRICE_5_GB) == 1 && rec.Cmp(PRICE_1_MONTH_HOME) == -1):
		voucherPayload.CouponSize = SIZE_5_GB
		tierDescription = PRICE_5_GB_DESC
	case rec.Cmp(PRICE_1_MONTH_HOME) == 0 || rec.Cmp(PRICE_1_MONTH_HOME) == 1:
		voucherPayload.CouponSize = SIZE_1_MONTH_HOME
		tierDescription = PRICE_1_MONTH_HOME_DESC
	case rec.Cmp(PRICE_1_MONTH_BUSINESS) == 0 || rec.Cmp(PRICE_1_MONTH_BUSINESS) == 1:
		voucherPayload.CouponSize = SIZE_1_MONTH_BUSINESS
		tierDescription = PRICE_1_MONTH_BUSINESS_DESC
	default:
		h.logg.Info("generate voucher skipped, unrecognized amount", "amount", rec)
		return nil
	}
	voucherPayload.Amount = formatAmount(rec)
	h.logg.Debug("generate voucher", "amount", voucherPayload.Amount, "size", voucherPayload.CouponSize, "tier", tierDescription)

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

	notifyResp, err := h.nClient.SendNotification(ctx, notify.NotifyPayload{
		SenderAddress: voucherPayload.SenderAddress,
		Code:          resp.Voucher,
		Size:          tierDescription,
	})
	if err != nil {
		h.logg.Error("failed to send notification", "error", err, "sender", voucherPayload.SenderAddress)
	} else {
		h.logg.Debug("notification sent successfully", "success", notifyResp.Success, "message", notifyResp.Message, "sender", voucherPayload.SenderAddress, "tier", tierDescription)
	}

	return nil
}

func formatAmount(dividend *big.Int) string {
	floatDividend := new(big.Float).SetInt(dividend)
	result := new(big.Float).Quo(floatDividend, AMOUNT_DIVISOR)
	return fmt.Sprintf("%.8f", result)
}
