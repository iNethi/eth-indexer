package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/grassrootseconomics/eth-indexer/internal/api"
	"github.com/grassrootseconomics/eth-indexer/internal/cache"
	"github.com/grassrootseconomics/eth-indexer/internal/handler"
	"github.com/grassrootseconomics/eth-indexer/internal/store"
	"github.com/grassrootseconomics/eth-indexer/internal/sub"
	"github.com/grassrootseconomics/eth-indexer/internal/telegram"
	"github.com/grassrootseconomics/eth-indexer/internal/util"
	"github.com/grassrootseconomics/ethutils"
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

	lo = util.InitLogger()
	ko = util.InitConfig(lo, confFlag)

	lo.Info("starting eth indexer", "build", build)
}

func main() {
	var wg sync.WaitGroup
	ctx, stop := notifyShutdown()

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

	cache := cache.New()

	chainProvider := ethutils.NewProvider(
		ko.MustString("chain.rpc_endpoint"),
		ko.MustInt64("chain.chainid"),
	)

	telegram := telegram.New(telegram.TelegramOpts{
		BotToken:            ko.MustString("telegram.bot_token"),
		NotificationChannel: ko.MustInt64("telegram.notification_channel"),
	})

	handlerContainer := handler.NewHandler(handler.HandlerOpts{
		Store:         store,
		Cache:         cache,
		ChainProvider: chainProvider,
		Telegram:      telegram,
		Logg:          lo,
	})

	router := bootstrapRouter(handlerContainer)

	jetStreamSub, err := sub.NewJetStreamSub(sub.JetStreamOpts{
		Logg:        lo,
		Router:      router,
		Endpoint:    ko.MustString("jetstream.endpoint"),
		JetStreamID: ko.MustString("jetstream.id"),
	})
	if err != nil {
		lo.Error("could not initialize jetstream sub", "error", err)
		os.Exit(1)
	}

	apiServer := &http.Server{
		Addr:    ko.MustString("api.address"),
		Handler: api.New(),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		jetStreamSub.Process()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		lo.Info("metrics and stats server starting", "address", ko.MustString("api.address"))
		if err := apiServer.ListenAndServe(); err != http.ErrServerClosed {
			lo.Error("failed to start API server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	lo.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultGracefulShutdownPeriod)

	wg.Add(1)
	go func() {
		defer wg.Done()
		jetStreamSub.Close()
		store.Close()
		apiServer.Shutdown(shutdownCtx)
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
