package postgresdb

import (
	"database/sql"
	"fmt"
	"github.com/siraj18/balance-service-new/internal/models"
	"math"
	"strings"
)

var ErrorUserNotFound = fmt.Errorf("user not found")
var ErrorNotEnoughMoney = fmt.Errorf("not enough money")
var ErrorInvalidInput = fmt.Errorf("invalid type for uid")
var ErrorNegativeAmount = fmt.Errorf("negative amount")

const (
	operationAddMoney           = "adding money"
	operationWithdrawMoney      = "withdrawal of money"
	operationTransferMoney      = "transfer money"
	operationReserveMoney       = "reserve money"
	operationReturnReserveMoney = "return reserve money"
)

func (rep *BalanceRepository) createUserBalance(uid string, tx *sql.Tx) error {
	_, err := tx.Exec(addUserSql, uid)

	return err
}

func (rep *BalanceRepository) ChangeBalance(uid string, money float64) (*models.User, error) {
	tx, err := rep.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var user models.User

	if err := tx.QueryRow(getUserSql, uid).Scan(&user.Id, &user.Balance); err != nil {
		if err == sql.ErrNoRows {
			err = rep.createUserBalance(uid, tx)
			if err != nil {
				return nil, err
			}

		} else {
			return nil, err
		}
	}

	if money < 0 {
		if user.Balance < math.Abs(money) {
			return nil, ErrorNotEnoughMoney
		}
	}

	row := tx.QueryRow(updateUserBalanceSql, uid, money)
	if err = row.Scan(&user.Id, &user.Balance); err != nil {
		return nil, err
	}

	if money >= 0 {
		if err = rep.addTransaction(&uid, nil, operationAddMoney, money, tx); err != nil {
			return nil, err
		}
	} else {
		if err = rep.addTransaction(nil, &uid, operationWithdrawMoney, money, tx); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &user, nil
}

func (rep *BalanceRepository) GetBalance(uid string) (*models.User, error) {
	var user models.User

	if err := rep.db.Get(&user, getUserSql, uid); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorUserNotFound
		}

		if strings.Contains(err.Error(), "ERROR: invalid input syntax for type uuid:") {
			return nil, ErrorInvalidInput
		}

		return nil, fmt.Errorf("error when get user balance: %w", err)
	}

	return &user, nil

}

func (rep *BalanceRepository) TransferBalance(fromUid string, toUid string, money float64) error {
	tx, err := rep.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if money < 0 {
		return ErrorNegativeAmount
	}

	var empty interface{}
	err = tx.QueryRow(updateUserBalanceSql, toUid, money).Scan(&empty, &empty)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrorUserNotFound
		}

		return err
	}

	err = tx.QueryRow(updateUserBalanceSql, fromUid, -money).Scan(&empty, &empty)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrorUserNotFound
		} else if strings.Contains(err.Error(), "users_balance_check") {
			return ErrorNotEnoughMoney
		}

		return err
	}

	if err = rep.addTransaction(&toUid, &fromUid, operationTransferMoney, money, tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
