package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/grassrootseconomics/celo-indexer/internal/handler"
	"github.com/grassrootseconomics/celo-indexer/internal/store"
	"github.com/grassrootseconomics/celo-indexer/internal/sub"
	"github.com/knadh/koanf/v2"
)

const defaultGracefulShutdownPeriod = time.Second * 20

var (
	build = "dev"

	confFlag             string
	migrationsFolderFlag string
	queriesFlag          string

	lo *slog.Logger
	ko *koanf.Koanf
)

func init() {
	flag.StringVar(&confFlag, "config", "config.toml", "Config file location")
	flag.StringVar(&migrationsFolderFlag, "migrations", "migrations/", "Migrations folder location")
	flag.StringVar(&queriesFlag, "queries", "queries.sql", "Queries file location")
	flag.Parse()

	lo = initLogger()
	ko = initConfig()

	lo.Info("starting celo indexer", "build", build)
}

func main() {
	var wg sync.WaitGroup
	ctx, stop := notifyShutdown()

	// chain, err := chain.NewChainProvider(chain.ChainOpts{
	// 	RPCEndpoint: ko.MustString("chain.rpc_endpoint"),
	// 	ChainID:     ko.MustInt64("chain.chainid"),
	// })
	// if err != nil {
	// 	lo.Error("chain provider bootstrap failed", "error", err)
	// 	os.Exit(1)
	// }
	// cache := cache.NewCache()

	// lo.Info("starting cache bootstrap this may take a few minutes")
	// if err := chain.BootstrapCache(ko.MustStrings("bootstrap.ge_registries"), cache); err != nil {
	// 	lo.Error("cache bootstrap failed", "error", err)
	// 	os.Exit(1)
	// }
	// lo.Info("cache bootstrap completed successfully", "cache_size", cache.Size())

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

	handler := handler.NewHandler(handler.HandlerOpts{
		Store: store,
		// Cache: cache,
	})

	jetStreamSub, err := sub.NewJetStreamSub(sub.JetStreamOpts{
		Logg:        lo,
		Store:       store,
		Handler:     handler,
		Endpoint:    ko.MustString("jetstream.endpoint"),
		JetStreamID: ko.MustString("jetstream.id"),
	})
	if err != nil {
		lo.Error("could not initialize jetstream sub", "error", err)
		os.Exit(1)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		jetStreamSub.Process()
	}()

	<-ctx.Done()
	lo.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultGracefulShutdownPeriod)

	wg.Add(1)
	go func() {
		defer wg.Done()
		jetStreamSub.Close()
	}()

	go func() {
		wg.Wait()
		stop()
		cancel()
		os.Exit(0)
	}()

	<-shutdownCtx.Done()
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		stop()
		cancel()
		lo.Error("graceful shutdown period exceeded, forcefully shutting down")
	}
	os.Exit(1)
}

func notifyShutdown() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
}
