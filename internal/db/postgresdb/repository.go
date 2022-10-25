package postgresdb

import "github.com/jmoiron/sqlx"

type BalanceRepository struct {
	db *sqlx.DB
}

func NewSqlRepository(db *sqlx.DB) (*BalanceRepository, error) {
	rep := &BalanceRepository{db}

	if err := rep.init(); err != nil {
		return nil, err
	}

	return rep, nil
}

func (rep *BalanceRepository) init() error {
	_, err := rep.db.Exec(initSchema)

	if err != nil {
		return err
	}

	return nil
}

func (rep *BalanceRepository) Close() {
	rep.db.Close()
}
