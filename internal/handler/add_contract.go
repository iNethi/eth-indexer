package handler

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/grassrootseconomics/ethutils"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

var (
	nameGetter        = w3.MustNewFunc("name()", "string")
	symbolGetter      = w3.MustNewFunc("symbol()", "string")
	decimalsGetter    = w3.MustNewFunc("decimals()", "uint8")
	sinkAddressGetter = w3.MustNewFunc("sinkAddress()", "address")
)

func (h *Handler) AddToken(ctx context.Context, event event.Event) error {
	if h.cache.Get(event.ContractAddress) {
		return nil
	}

	var (
		tokenName     string
		tokenSymbol   string
		tokenDecimals uint8
		sinkAddress   common.Address

		batchErr w3.CallErrors
	)

	contractAddress := w3.A(event.ContractAddress)

	if err := h.chainProvider.Client.CallCtx(
		ctx,
		eth.CallFunc(contractAddress, nameGetter).Returns(&tokenName),
		eth.CallFunc(contractAddress, symbolGetter).Returns(&tokenSymbol),
		eth.CallFunc(contractAddress, decimalsGetter).Returns(&tokenDecimals),
	); errors.As(err, &batchErr) {
		return batchErr
	} else if err != nil {
		return err
	}

	if err := h.chainProvider.Client.CallCtx(
		ctx,
		eth.CallFunc(contractAddress, decimalsGetter).Returns(&tokenDecimals),
	); err != nil {
		// This will most likely revert if the contract does not have a sinkAddress
		// Instead of handling the error we just ignore it and set the value to 0
		sinkAddress = ethutils.ZeroAddress
	}

	return h.store.InsertToken(ctx, event.ContractAddress, tokenName, tokenSymbol, tokenDecimals, sinkAddress.Hex())
}

func (h *Handler) AddPool(ctx context.Context, event event.Event) error {
	if h.cache.Get(event.ContractAddress) {
		return nil
	}

	var (
		tokenName   string
		tokenSymbol string
	)

	contractAddress := w3.A(event.ContractAddress)

	if err := h.chainProvider.Client.CallCtx(
		ctx,
		eth.CallFunc(contractAddress, nameGetter).Returns(&tokenName),
		eth.CallFunc(contractAddress, symbolGetter).Returns(&tokenSymbol),
	); err != nil {
		return err
	}

	return h.store.InsertPool(ctx, event.ContractAddress, tokenName, tokenSymbol)
}
