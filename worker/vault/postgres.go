package vault

import (
	"context"
	_ "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"gitlab.com/logiq.one/agenda_3dru_bot/config"
)

type Postgreser struct {
	DbPool *pgxpool.Pool
}

func New(cfg *config.App) (*Postgreser, error) {
	dbPool, err := pgxpool.Connect(context.Background(), cfg.Database.ConnStr)

	if err != nil {
		return nil, err
	}

	return &Postgreser{dbPool}, nil
}
