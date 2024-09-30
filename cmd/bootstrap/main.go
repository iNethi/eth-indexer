package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/grassrootseconomics/eth-indexer/internal/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/knadh/koanf/v2"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/ethutils"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

type TokenArgs struct {
	ContractAddress string
	TokenName       string
	TokenSymbol     string
	TokenDecimals   uint8
}

const (
	insertTokenQuery = `INSERT INTO tokens(
		contract_address,
		token_name,
		token_symbol,
		token_decimals
	) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
)

var (
	build = "dev"

	confFlag string

	lo *slog.Logger
	ko *koanf.Koanf

	dbPool *pgxpool.Pool
)

func init() {
	flag.StringVar(&confFlag, "config", "config.toml", "Config file location")
	flag.Parse()

	lo = util.InitLogger()
	ko = util.InitConfig(lo, confFlag)

	lo.Info("starting GE indexer token bootstrapper", "build", build)
}

func main() {
	var (
		tokenRegistryGetter = w3.MustNewFunc("tokenRegistry()", "address")
		nameGetter          = w3.MustNewFunc("name()", "string")
		symbolGetter        = w3.MustNewFunc("symbol()", "string")
		decimalsGetter      = w3.MustNewFunc("decimals()", "uint8")
	)

	chainProvider := ethutils.NewProvider(ko.MustString("chain.rpc_endpoint"), ko.MustInt64("chain.chainid"))

	var err error
	dbPool, err = newPgStore()
	if err != nil {
		lo.Error("could not initialize postgres store", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	for _, registry := range ko.MustStrings("bootstrap.ge_registries") {
		registryMap, err := chainProvider.RegistryMap(ctx, ethutils.HexToAddress(registry))
		if err != nil {
			lo.Error("could not fetch registry", "error", err)
			os.Exit(1)
		}

		if tokenIndex := registryMap[ethutils.TokenIndex]; tokenIndex != ethutils.ZeroAddress {
			tokenIndexIter, err := chainProvider.NewBatchIterator(ctx, tokenIndex)
			if err != nil {
				lo.Error("could not create token index iter", "error", err)
				os.Exit(1)
			}

			for {
				batch, err := tokenIndexIter.Next(ctx)
				if err != nil {
					lo.Error("error fetching next token index batch", "error", err)
					os.Exit(1)
				}
				if batch == nil {
					break
				}
				lo.Debug("index batch", "index", tokenIndex.Hex(), "size", len(batch))
				for _, address := range batch {
					if address != ethutils.ZeroAddress {
						var (
							tokenName     string
							tokenSymbol   string
							tokenDecimals uint8
						)

						err := chainProvider.Client.CallCtx(
							ctx,
							eth.CallFunc(address, nameGetter).Returns(&tokenName),
							eth.CallFunc(address, symbolGetter).Returns(&tokenSymbol),
							eth.CallFunc(address, decimalsGetter).Returns(&tokenDecimals),
						)
						if err != nil {
							lo.Error("error fetching token details", "error", err)
							os.Exit(1)
						}

						if err := insertToken(ctx, TokenArgs{
							ContractAddress: address.Hex(),
							TokenName:       tokenName,
							TokenSymbol:     tokenSymbol,
							TokenDecimals:   tokenDecimals,
						}); err != nil {
							lo.Error("pg insert error", "error", err)
							os.Exit(1)
						}
					}
				}
			}
		}

		if poolIndex := registryMap[ethutils.PoolIndex]; poolIndex != ethutils.ZeroAddress {
			poolIndexIter, err := chainProvider.NewBatchIterator(ctx, poolIndex)
			if err != nil {
				lo.Error("cache could create pool index iter", "error", err)
				os.Exit(1)
			}

			for {
				batch, err := poolIndexIter.Next(ctx)
				if err != nil {
					lo.Error("error fetching next pool index batch", "error", err)
					os.Exit(1)
				}
				if batch == nil {
					break
				}
				lo.Debug("index batch", "index", poolIndex.Hex(), "size", len(batch))
				for _, address := range batch {
					var poolTokenIndex common.Address
					err := chainProvider.Client.CallCtx(
						ctx,
						eth.CallFunc(address, tokenRegistryGetter).Returns(&poolTokenIndex),
					)
					if err != nil {
						lo.Error("error fetching pool token index and/or quoter", "error", err)
						os.Exit(1)
					}
					if poolTokenIndex != ethutils.ZeroAddress {
						poolTokenIndexIter, err := chainProvider.NewBatchIterator(ctx, poolTokenIndex)
						if err != nil {
							lo.Error("error creating pool token index iter", "error", err)
							os.Exit(1)
						}

						for {
							batch, err := poolTokenIndexIter.Next(ctx)
							if err != nil {
								lo.Error("error fetching next pool token index batch", "error", err)
								os.Exit(1)
							}
							if batch == nil {
								break
							}
							lo.Debug("index batch", "index", poolTokenIndex.Hex(), "size", len(batch))
							for _, address := range batch {
								if address != ethutils.ZeroAddress {
									var (
										tokenName     string
										tokenSymbol   string
										tokenDecimals uint8
									)

									err := chainProvider.Client.CallCtx(
										ctx,
										eth.CallFunc(address, nameGetter).Returns(&tokenName),
										eth.CallFunc(address, symbolGetter).Returns(&tokenSymbol),
										eth.CallFunc(address, decimalsGetter).Returns(&tokenDecimals),
									)
									if err != nil {
										lo.Error("error fetching token details", "error", err)
										os.Exit(1)
									}

									if err := insertToken(ctx, TokenArgs{
										ContractAddress: address.Hex(),
										TokenName:       tokenName,
										TokenSymbol:     tokenSymbol,
										TokenDecimals:   tokenDecimals,
									}); err != nil {
										lo.Error("pg insert error", "error", err)
										os.Exit(1)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	lo.Info("tokens bootstrap complete")
}

func newPgStore() (*pgxpool.Pool, error) {
	parsedConfig, err := pgxpool.ParseConfig(ko.MustString("postgres.dsn"))
	if err != nil {
		return nil, err
	}

	dbPool, err := pgxpool.NewWithConfig(context.Background(), parsedConfig)
	if err != nil {
		return nil, err
	}

	return dbPool, nil
}

func insertToken(ctx context.Context, insertArgs TokenArgs) error {
	_, err := dbPool.Exec(
		ctx,
		insertTokenQuery,
		insertArgs.ContractAddress,
		insertArgs.TokenName,
		insertArgs.TokenSymbol,
		insertArgs.TokenDecimals,
	)
	if err != nil {
		return err
	}

	return nil
}
