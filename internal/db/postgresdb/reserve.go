package postgresdb

import (
	"database/sql"
	"fmt"
	"github.com/siraj18/balance-service-new/internal/models"
	"strings"
	"time"
)

var ErrorReserveNotFound = fmt.Errorf("reserve not found")
var ErrorReserveAlreadyRecognized = fmt.Errorf("reserve already recognized")
var ErrorReserveAlreadyDeReserved = fmt.Errorf("reserve already de-reserved")

const (
	statusReserveMoney    = "reserved"
	statusDeReservedMoney = "de-reserved"
	statusRecognizedMoney = "recognized"
)

func (rep *BalanceRepository) addReserve(userId, serviceId, orderId, status string, amount float64, tx *sql.Tx) error {
	_, err := tx.Exec(addReserveSql, userId, serviceId, orderId, amount, status, time.Now())

	return err
}

func (rep *BalanceRepository) ReserveMoney(userId, serviceId, orderId string, amount float64) error {
	tx, err := rep.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var user models.User

	if amount < 0 {
		return ErrorNegativeAmount
	}

	if err := tx.QueryRow(getUserSql, userId).Scan(&user.Id, &user.Balance); err != nil {
		if err == sql.ErrNoRows {
			return ErrorUserNotFound
		}

		if strings.Contains(err.Error(), "ERROR: invalid input syntax for type uuid:") {
			return ErrorInvalidInput
		}

		return err
	}

	if user.Balance < amount {
		return ErrorNotEnoughMoney
	}

	var empty interface{}
	if err = tx.QueryRow(updateUserBalanceSql, userId, -amount).Scan(&empty, &empty); err != nil {
		return err
	}

	if err = rep.addReserve(userId, serviceId, orderId, statusReserveMoney, amount, tx); err != nil {
		return err
	}

	if err = rep.addTransaction(nil, &userId, operationReserveMoney, amount, tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (rep *BalanceRepository) RecognizedMoney(userId, serviceId, orderId string, amount float64) error {
	tx, err := rep.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var reserve models.Reserve

	if err := rep.db.Get(&reserve, getReserveSql, userId, serviceId, orderId, amount); err != nil {
		if err == sql.ErrNoRows {
			return ErrorReserveNotFound
		}

		if strings.Contains(err.Error(), "ERROR: invalid input syntax for type uuid:") {
			return ErrorInvalidInput
		}

		return fmt.Errorf("error when get reserve: %w", err)
	}

	if reserve.Status != statusReserveMoney {
		if reserve.Status == statusRecognizedMoney {
			return ErrorReserveAlreadyRecognized
		}

		return ErrorReserveAlreadyDeReserved
	}

	if err = tx.QueryRow(updateReserveStatus, reserve.Id, statusRecognizedMoney, time.Now()).Err(); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (rep *BalanceRepository) DeReserveMoney(userId, serviceId, orderId string, amount float64) error {
	tx, err := rep.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var reserve models.Reserve

	if err := rep.db.Get(&reserve, getReserveSql, userId, serviceId, orderId, amount); err != nil {
		if err == sql.ErrNoRows {
			return ErrorReserveNotFound
		}

		if strings.Contains(err.Error(), "ERROR: invalid input syntax for type uuid:") {
			return ErrorInvalidInput
		}

		return fmt.Errorf("error when get reserve: %w", err)
	}

	if reserve.Status != statusReserveMoney {
		if reserve.Status == statusRecognizedMoney {
			return ErrorReserveAlreadyRecognized
		}

		return ErrorReserveAlreadyDeReserved
	}

	var empty interface{}
	if err = tx.QueryRow(updateUserBalanceSql, userId, reserve.Amount).Scan(&empty, &empty); err != nil {
		if err == sql.ErrNoRows {
			return ErrorUserNotFound
		}

		return err
	}

	if err = rep.addTransaction(&userId, nil, operationReturnReserveMoney, amount, tx); err != nil {
		return err
	}

	if err = tx.QueryRow(updateReserveStatus, reserve.Id, statusDeReservedMoney, nil).Err(); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (rep *BalanceRepository) GetReserves(year, month int) (*[]models.Reserve, error) {
	reserves := []models.Reserve{}

	err := rep.db.Select(&reserves, getReserveForReportSql, statusRecognizedMoney, year, month)
	if err != nil {
		return nil, err
	}

	return &reserves, nil
}
