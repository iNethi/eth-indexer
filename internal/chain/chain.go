package chain

import (
	"context"
	"math/big"
	"net/http"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/rpc"
	"github.com/grassrootseconomics/celo-indexer/internal/cache"
	"github.com/grassrootseconomics/celoutils/v3"
	"github.com/grassrootseconomics/w3-celo"
	"github.com/grassrootseconomics/w3-celo/module/eth"
	"github.com/grassrootseconomics/w3-celo/w3types"
)

type (
	ChainOpts struct {
		RPCEndpoint string
		ChainID     int64
	}

	Chain struct {
		Provider *celoutils.Provider
	}
)

const bootstrapTimeout = 2 * time.Minute

var (
	entryCountFunc = w3.MustNewFunc("entryCount()", "uint256")
	entrySig       = w3.MustNewFunc("entry(uint256 _idx)", "address")
)

func NewChainProvider(o ChainOpts) (*Chain, error) {
	customRPCClient, err := lowTimeoutRPCClient(o.RPCEndpoint)
	if err != nil {
		return nil, err
	}

	chainProvider := celoutils.NewProvider(
		o.RPCEndpoint,
		o.ChainID,
		celoutils.WithClient(customRPCClient),
	)

	return &Chain{
		Provider: chainProvider,
	}, nil
}

func lowTimeoutRPCClient(rpcEndpoint string) (*w3.Client, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	rpcClient, err := rpc.DialHTTPWithClient(
		rpcEndpoint,
		httpClient,
	)
	if err != nil {
		return nil, err
	}

	return w3.NewClient(rpcClient), nil
}

func (c *Chain) BootstrapCache(registries []string, cache *cache.Cache) error {
	var ZeroAddress common.Address
	for _, registry := range registries {
		ctx, cancel := context.WithTimeout(context.Background(), bootstrapTimeout)
		defer cancel()

		registryMap, err := c.Provider.RegistryMap(ctx, w3.A(registry))
		if err != nil {
			return err
		}

		if accountIndex := registryMap[celoutils.AccountIndex]; accountIndex != ZeroAddress {
			if err := c.getAllAccountsFromAccountIndex(ctx, accountIndex, cache); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Chain) getAllAccountsFromAccountIndex(ctx context.Context, accountIndex common.Address, cache *cache.Cache) error {
	var accountIndexEntryCount big.Int

	if err := c.Provider.Client.CallCtx(
		ctx,
		eth.CallFunc(accountIndex, entryCountFunc).Returns(&accountIndexEntryCount),
	); err != nil {
		return err
	}

	// TODO: Temporary skip custodial bootsrap, load from DB
	if accountIndexEntryCount.Int64() > 1000 {
		return nil
	}

	calls := make([]w3types.RPCCaller, accountIndexEntryCount.Int64())
	accountAddresses := make([]common.Address, accountIndexEntryCount.Int64())

	for i := 0; i < int(accountIndexEntryCount.Int64()); i++ {
		calls[i] = eth.CallFunc(accountIndex, entrySig, new(big.Int).SetInt64(int64(i))).Returns(&accountAddresses[i])
	}

	if err := c.Provider.Client.CallCtx(ctx, calls...); err != nil {
		return err
	}

	for _, address := range accountAddresses {
		cache.Add(address.Hex())
	}

	return nil
}
