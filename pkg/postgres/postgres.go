package postgres

import (
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func NewDb(conStr string, retries int) (*sqlx.DB, error) {
	var err error
	var db *sqlx.DB

	for i := 0; i < retries; i++ {
		db, err = sqlx.Connect("pgx", conStr)
		if err != nil {
			logrus.Error("unable to connect db:%w", err)
			time.Sleep(time.Second * 2)
		} else if err = db.Ping(); err != nil {
			logrus.Error("an error occurred when ping database: %w", err)
			time.Sleep(time.Second * 2)
		} else {
			err = nil
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return db, nil
}
