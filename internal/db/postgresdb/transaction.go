package postgresdb

import (
	"database/sql"
	"fmt"
	"github.com/siraj18/balance-service-new/internal/models"
	"strings"
	"time"
)

const (
	sortDateAsc   = "date_asc"
	sortDateDesc  = "date_desc"
	sortMoneyAsc  = "money_asc"
	sortMoneyDesc = "money_desc"
)

var ErrorInvalidSortParameters = fmt.Errorf("invalid sort parameters")

func (rep *BalanceRepository) addTransaction(toId, fromId *string, operation string, money float64, tx *sql.Tx) error {
	_, err := tx.Exec(addTransactionsSql, toId, fromId, money, operation, time.Now())

	return err
}

func (rep *BalanceRepository) GetAllTransactions(id string, sortType string, limit int, page int) (*[]models.Transaction, error) {
	finalSql := ""

	if limit < 0 || page < 0 {
		return nil, ErrorInvalidSortParameters
	}

	switch strings.ToLower(sortType) {
	case sortDateAsc:
		finalSql = fmt.Sprintf(getAllTransactionsSql, "ORDER BY created_at ASC")
	case sortDateDesc:
		finalSql = fmt.Sprintf(getAllTransactionsSql, "ORDER BY created_at DESC")
	case sortMoneyAsc:
		finalSql = fmt.Sprintf(getAllTransactionsSql, "ORDER BY money ASC")
	case sortMoneyDesc:
		finalSql = fmt.Sprintf(getAllTransactionsSql, "ORDER BY money DESC")
	default:
		finalSql = getAllTransactionsSql
	}

	transactions := []models.Transaction{}

	offset := (page - 1) * limit

	err := rep.db.Select(&transactions, finalSql, id, limit, offset)
	if err != nil {
		if strings.Contains(err.Error(), "ERROR: invalid input syntax for type uuid:") {
			return nil, ErrorInvalidInput
		}

		return nil, err
	}

	return &transactions, nil
}
