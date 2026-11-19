package database

import (
	"context"
	"fmt"
	"time"

	"backoffice-service/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PostgresDB struct {
	Master  *pgxpool.Pool
	Replica *pgxpool.Pool
}

func NewPostgresClient(ctx context.Context, cfg config.DBConfig, log *zap.Logger) (*PostgresDB, error) {
	// Configure Master Connection Pool
	masterConfig, err := pgxpool.ParseConfig(cfg.MasterDSN)
	if err != nil {
		return nil, fmt.Errorf("unable to parse master dsn: %w", err)
	}

	masterConfig.MaxConns = cfg.MaxConns
	masterConfig.MinConns = cfg.MinConns
	masterConfig.MaxConnIdleTime = cfg.MaxConnIdle
	masterConfig.MaxConnLifetime = cfg.MaxConnLife

	masterPool, err := pgxpool.NewWithConfig(ctx, masterConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to master postgres: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := masterPool.Ping(pingCtx); err != nil {
		log.Warn("master postgres ping warning (might be offline during setup)", zap.Error(err))
	} else {
		log.Info("Connected to PostgreSQL Master successfully")
	}

	// Configure Replica Connection Pool
	var replicaPool *pgxpool.Pool
	if cfg.ReplicaDSN != "" && cfg.ReplicaDSN != cfg.MasterDSN {
		replicaConfig, err := pgxpool.ParseConfig(cfg.ReplicaDSN)
		if err == nil {
			replicaConfig.MaxConns = cfg.MaxConns
			replicaConfig.MinConns = cfg.MinConns
			replicaConfig.MaxConnIdleTime = cfg.MaxConnIdle
			replicaConfig.MaxConnLifetime = cfg.MaxConnLife

			rPool, err := pgxpool.NewWithConfig(ctx, replicaConfig)
			if err == nil && rPool.Ping(pingCtx) == nil {
				replicaPool = rPool
				log.Info("Connected to PostgreSQL Replica successfully")
			}
		}
	}

	if replicaPool == nil {
		replicaPool = masterPool
		log.Info("Replica pool defaulting to master pool")
	}

	return &PostgresDB{
		Master:  masterPool,
		Replica: replicaPool,
	}, nil
}

func (db *PostgresDB) Close() {
	if db.Master != nil {
		db.Master.Close()
	}
	if db.Replica != nil && db.Replica != db.Master {
		db.Replica.Close()
	}
}

