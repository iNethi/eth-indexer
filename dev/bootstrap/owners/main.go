package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/grassrootseconomics/eth-indexer/internal/store"
	"github.com/grassrootseconomics/eth-indexer/internal/util"
	"github.com/grassrootseconomics/ethutils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/knadh/koanf/v2"
	"github.com/lmittmann/w3"
)

const (
	insertOwnerQuery = `INSERT INTO ownership_change(
		new_owner
		contract_address
	) VALUES ($1, $2) ON CONFLICT DO NOTHING`

	getTokens = `SELECT contract_address FROM tokens`

	getPools = `SELECT contract_address FROM pools`
)

var (
	build = "dev"

	confFlag             string
	migrationsFolderFlag string
	queriesFlag          string

	lo *slog.Logger
	ko *koanf.Koanf

	dbPool *pgxpool.Pool
)

func init() {
	flag.StringVar(&confFlag, "config", "config.toml", "Config file location")
	flag.StringVar(&migrationsFolderFlag, "migrations", "migrations/", "Migrations folder location")
	flag.StringVar(&queriesFlag, "queries", "queries.sql", "Queries file location")
	flag.Parse()

	lo = util.InitLogger()
	ko = util.InitConfig(lo, confFlag)

	lo.Info("starting owners bootstrapper", "build", build)
}

func main() {
	var ownerGetter = w3.MustNewFunc("owner()", "address")

	chainProvider := ethutils.NewProvider(ko.MustString("chain.rpc_endpoint"), ko.MustInt64("chain.chainid"))

	var err error
	dbPool, err = newPgStore()
	if err != nil {
		lo.Error("could not initialize postgres store", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	// TODO: get all tokens and pools

}

func newPgStore() (*pgxpool.Pool, error) {
	store, err := store.NewPgStore(store.PgOpts{
		Logg:                 lo,
		DSN:                  ko.MustString("postgres.dsn"),
		MigrationsFolderPath: migrationsFolderFlag,
		QueriesFolderPath:    queriesFlag,
	})
	if err != nil {
		lo.Error("could not initialize postgres store", "error", err)
		os.Exit(1)
	}

	return store.Pool(), nil
}

func insertOwnershipChange(ctx context.Context, owner string, contractAddress string) error {
	_, err := dbPool.Exec(
		ctx,
		insertOwnerQuery,
		owner,
		contractAddress,
	)
	if err != nil {
		return err
	}

	return nil
}
