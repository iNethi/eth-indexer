package store

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"github.com/knadh/goyesql/v2"
)

type (
	PgOpts struct {
		Logg                 *slog.Logger
		DSN                  string
		MigrationsFolderPath string
		QueriesFolderPath    string
	}

	Pg struct {
		logg    *slog.Logger
		db      *pgxpool.Pool
		queries *queries
	}

	queries struct {
		InsertTx               string `query:"insert-tx"`
		InsertTokenTransfer    string `query:"insert-token-transfer"`
		InsertTokenMint        string `query:"insert-token-mint"`
		InsertTokenBurn        string `query:"insert-token-burn"`
		InsertFaucetGive       string `query:"insert-faucet-give"`
		InsertPoolSwap         string `query:"insert-pool-swap"`
		InsertPoolDeposit      string `query:"insert-pool-deposit"`
		InsertPriceQuoteUpdate string `query:"insert-price-quote-update"`
		CheckAddressExists     string `query:"address-exists"`
	}
)

func NewPgStore(o PgOpts) (Store, error) {
	parsedConfig, err := pgxpool.ParseConfig(o.DSN)
	if err != nil {
		return nil, err
	}

	dbPool, err := pgxpool.NewWithConfig(context.Background(), parsedConfig)
	if err != nil {
		return nil, err
	}

	queries, err := loadQueries(o.QueriesFolderPath)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(context.Background(), dbPool, o.MigrationsFolderPath); err != nil {
		return nil, err
	}
	o.Logg.Info("migrations ran successfully")

	return &Pg{
		logg:    o.Logg,
		db:      dbPool,
		queries: queries,
	}, nil
}

func (pg *Pg) InsertTokenTransfer(ctx context.Context, eventPayload event.Event) error {
	return pg.executeTransaction(ctx, func(tx pgx.Tx) error {
		// var addressExists bool

		// if err := tx.QueryRow(
		// 	ctx,
		// 	pg.queries.CheckAddressExists,
		// 	eventPayload.Payload["from"].(string),
		// 	eventPayload.Payload["to"].(string),
		// ).Scan(&addressExists); err != nil {
		// 	return err
		// }

		// if !addressExists {
		// 	return nil
		// }
		// pg.logg.Debug("transfer address exists", "hash", eventPayload.TxHash)

		txID, err := pg.insertTx(ctx, tx, eventPayload)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			pg.queries.InsertTokenTransfer,
			txID,
			eventPayload.Payload["from"].(string),
			eventPayload.Payload["to"].(string),
			eventPayload.Payload["value"].(string),
		)
		return err
	})
}

func (pg *Pg) InsertTokenMint(ctx context.Context, eventPayload event.Event) error {
	return pg.executeTransaction(ctx, func(tx pgx.Tx) error {
		txID, err := pg.insertTx(ctx, tx, eventPayload)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			pg.queries.InsertTokenMint,
			txID,
			eventPayload.Payload["tokenMinter"].(string),
			eventPayload.Payload["to"].(string),
			eventPayload.Payload["value"].(string),
		)
		return err
	})
}

func (pg *Pg) InsertTokenBurn(ctx context.Context, eventPayload event.Event) error {
	return pg.executeTransaction(ctx, func(tx pgx.Tx) error {
		txID, err := pg.insertTx(ctx, tx, eventPayload)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			pg.queries.InsertTokenBurn,
			txID,
			eventPayload.Payload["tokenBurner"].(string),
			eventPayload.Payload["value"].(string),
		)
		return err
	})
}

func (pg *Pg) InsertFaucetGive(ctx context.Context, eventPayload event.Event) error {
	return pg.executeTransaction(ctx, func(tx pgx.Tx) error {
		txID, err := pg.insertTx(ctx, tx, eventPayload)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			pg.queries.InsertFaucetGive,
			txID,
			eventPayload.Payload["token"].(string),
			eventPayload.Payload["recipient"].(string),
			eventPayload.Payload["amount"].(string),
		)
		return err
	})
}

func (pg *Pg) InsertPoolSwap(ctx context.Context, eventPayload event.Event) error {
	return pg.executeTransaction(ctx, func(tx pgx.Tx) error {
		txID, err := pg.insertTx(ctx, tx, eventPayload)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			pg.queries.InsertPoolSwap,
			txID,
			eventPayload.Payload["initiator"].(string),
			eventPayload.Payload["tokenIn"].(string),
			eventPayload.Payload["tokenOut"].(string),
			eventPayload.Payload["amountIn"].(string),
			eventPayload.Payload["amountOut"].(string),
			eventPayload.Payload["fee"].(string),
		)
		return err
	})
}

func (pg *Pg) InsertPoolDeposit(ctx context.Context, eventPayload event.Event) error {
	return pg.executeTransaction(ctx, func(tx pgx.Tx) error {
		txID, err := pg.insertTx(ctx, tx, eventPayload)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			pg.queries.InsertPoolDeposit,
			txID,
			eventPayload.Payload["initiator"].(string),
			eventPayload.Payload["tokenIn"].(string),
			eventPayload.Payload["amountIn"].(string),
		)
		return err
	})
}

func (pg *Pg) InsertPriceQuoteUpdate(ctx context.Context, eventPayload event.Event) error {
	return pg.executeTransaction(ctx, func(tx pgx.Tx) error {
		txID, err := pg.insertTx(ctx, tx, eventPayload)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			pg.queries.InsertPriceQuoteUpdate,
			txID,
			eventPayload.Payload["token"].(string),
			eventPayload.Payload["exchangeRate"].(string),
		)
		return err
	})
}

func (pg *Pg) insertTx(ctx context.Context, tx pgx.Tx, eventPayload event.Event) (int, error) {
	var txID int
	if err := tx.QueryRow(
		ctx,
		pg.queries.InsertTx,
		eventPayload.TxHash,
		eventPayload.Block,
		eventPayload.ContractAddress,
		time.Unix(int64(eventPayload.Timestamp), 0).UTC(),
		eventPayload.Success,
	).Scan(&txID); err != nil {
		return 0, err
	}
	return txID, nil
}

func (pg *Pg) executeTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := pg.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}

	return nil
}

func loadQueries(queriesPath string) (*queries, error) {
	parsedQueries, err := goyesql.ParseFile(queriesPath)
	if err != nil {
		return nil, err
	}

	loadedQueries := &queries{}

	if err := goyesql.ScanToStruct(loadedQueries, parsedQueries, nil); err != nil {
		return nil, fmt.Errorf("failed to scan queries %v", err)
	}

	return loadedQueries, nil
}

func runMigrations(ctx context.Context, dbPool *pgxpool.Pool, migrationsPath string) error {
	// Safety check
	// https://github.com/jackc/tern/issues/98
	envVars := []string{"REMOTE_DB_HOST", "REMOTE_DB_PORT", "REMOTE_DB_NAME", "REMOTE_DB_USER", "REMOTE_DB_PASSWORD"}
	for _, envVar := range envVars {
		_, exists := os.LookupEnv(envVar)
		if !exists {
			return fmt.Errorf("required env var %s not set", envVar)
		}
	}

	const migratorTimeout = 15 * time.Second

	ctx, cancel := context.WithTimeout(ctx, migratorTimeout)
	defer cancel()

	conn, err := dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	migrator, err := migrate.NewMigrator(ctx, conn.Conn(), "schema_version")
	if err != nil {
		return err
	}

	if err := migrator.LoadMigrations(os.DirFS(migrationsPath)); err != nil {
		return err
	}

	if err := migrator.Migrate(ctx); err != nil {
		return err
	}

	return nil
}
