package store

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/grassrootseconomics/celo-indexer/internal/event"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"github.com/knadh/goyesql/v2"
)

type (
	PgOpts struct {
		DSN                  string
		MigrationsFolderPath string
		QueriesFolderPath    string
		Logg                 *slog.Logger
	}

	Pg struct {
		db      *pgxpool.Pool
		queries *queries
		logg    *slog.Logger
	}

	queries struct {
		InsertTx            string `query:"insert-tx"`
		InsertTokenTransfer string `query:"insert-token-transfer"`
		InsertTokenMint     string `query:"insert-token-mint"`
		InsertPoolSwap      string `query:"insert-pool-swap"`
		InsertPoolDeposit   string `query:"insert-pool-deposit"`
	}
)

const (
	migratorTimeout = 5 * time.Second
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

	return &Pg{
		db:      dbPool,
		queries: queries,
		logg:    o.Logg,
	}, nil
}

func (pg *Pg) InsertTokenTransfer(ctx context.Context, eventPayload event.Event) error {
	tx, err := pg.db.Begin(ctx)
	if err != nil {
		pg.logg.Error("ERR0")
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	var (
		txID int
	)
	if err := tx.QueryRow(
		ctx,
		pg.queries.InsertTx,
		eventPayload.TxHash,
		eventPayload.Block,
		eventPayload.ContractAddress,
		time.Unix(eventPayload.Timestamp, 0).UTC(),
		eventPayload.Success,
	).Scan(&txID); err != nil {
		pg.logg.Error("ERR1")
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
	if err != nil {
		pg.logg.Error("ERR2")
		return err
	}

	return nil
}

func (pg *Pg) InsertTokenMint(ctx context.Context, eventPayload event.Event) error {
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

	var (
		txID int
	)
	if err := tx.QueryRow(
		ctx,
		pg.queries.InsertTx,
		eventPayload.TxHash,
		eventPayload.Block,
		eventPayload.ContractAddress,
		time.Unix(eventPayload.Timestamp, 0).UTC(),
		eventPayload.Success,
	).Scan(&txID); err != nil {
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
	if err != nil {
		return err
	}

	return nil
}

func (pg *Pg) InsertPoolSwap(ctx context.Context, eventPayload event.Event) error {
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

	var (
		txID int
	)
	if err := tx.QueryRow(
		ctx,
		pg.queries.InsertTx,
		eventPayload.TxHash,
		eventPayload.Block,
		eventPayload.ContractAddress,
		time.Unix(eventPayload.Timestamp, 0).UTC(),
		eventPayload.Success,
	).Scan(&txID); err != nil {
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
	if err != nil {
		return err
	}

	return nil
}

func (pg *Pg) InsertPoolDeposit(ctx context.Context, eventPayload event.Event) error {
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

	var (
		txID int
	)
	if err := tx.QueryRow(
		ctx,
		pg.queries.InsertTx,
		eventPayload.TxHash,
		eventPayload.Block,
		eventPayload.ContractAddress,
		time.Unix(eventPayload.Timestamp, 0).UTC(),
		eventPayload.Success,
	).Scan(&txID); err != nil {
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
	if err != nil {
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
