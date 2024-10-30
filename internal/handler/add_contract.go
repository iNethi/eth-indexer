package handler

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

var (
	nameGetter        = w3.MustNewFunc("name()", "string")
	symbolGetter      = w3.MustNewFunc("symbol()", "string")
	decimalsGetter    = w3.MustNewFunc("decimals()", "uint8")
	sinkAddressGetter = w3.MustNewFunc("sinkAddress", "address")
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
	)

	contractAddress := w3.A(event.ContractAddress)

	if err := h.chainProvider.Client.CallCtx(
		ctx,
		eth.CallFunc(contractAddress, nameGetter).Returns(&tokenName),
		eth.CallFunc(contractAddress, symbolGetter).Returns(&tokenSymbol),
		eth.CallFunc(contractAddress, decimalsGetter).Returns(&tokenDecimals),
		eth.CallFunc(contractAddress, sinkAddressGetter).Returns(&sinkAddress),
	); err != nil {
		return err
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
