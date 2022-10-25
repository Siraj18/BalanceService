package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/siraj18/balance-service-new/internal/db/postgresdb"
	"github.com/siraj18/balance-service-new/internal/handlers"
	"github.com/siraj18/balance-service-new/internal/handlers/mocks"
	"github.com/siraj18/balance-service-new/internal/models"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type handlerSuite struct {
	suite.Suite
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(handlerSuite))
}

func (t *handlerSuite) Test_getBalanceSuccess() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	userBalance := 50.0

	rep := mocks.NewMockRepository()
	rep.On("GetBalance", userId).Return(&models.User{
		Id:      userId,
		Balance: userBalance,
	}, nil)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	req, err := http.NewRequest("GET", testSrv.URL+"/balance/"+userId, nil)
	t.Nil(err)

	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	u := models.User{}
	json.NewDecoder(resp.Body).Decode(&u)

	t.Equal(http.StatusOK, resp.StatusCode)
	t.Equal(userId, u.Id)
	t.Equal(userBalance, u.Balance)
}

func (t *handlerSuite) Test_getBalanceUserNotFound() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"

	rep := mocks.NewMockRepository()
	rep.On("GetBalance", userId).Return(nil, postgresdb.ErrorUserNotFound)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	req, err := http.NewRequest("GET", testSrv.URL+"/balance/"+userId, nil)
	t.Nil(err)

	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	t.Equal(http.StatusNotFound, resp.StatusCode)
}

func (t *handlerSuite) Test_getBalanceSomeError() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"

	rep := mocks.NewMockRepository()
	rep.On("GetBalance", userId).Return(nil, fmt.Errorf("some error"))

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	req, err := http.NewRequest("GET", testSrv.URL+"/balance/"+userId, nil)
	t.Nil(err)

	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	t.Equal(http.StatusInternalServerError, resp.StatusCode)
}

func (t *handlerSuite) Test_addBalanceSuccess() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	userBalance := 50.0

	rep := mocks.NewMockRepository()
	rep.On("ChangeBalance", userId, userBalance).Return(&models.User{
		Id:      userId,
		Balance: userBalance,
	}, nil)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"id":    userId,
			"money": userBalance,
		},
	)
	t.Nil(err)

	req, err := http.NewRequest("POST", testSrv.URL+"/changeBalance", bytes.NewReader(body))
	t.Nil(err)

	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	u := models.User{}
	json.NewDecoder(resp.Body).Decode(&u)

	t.Equal(userId, u.Id)
	t.Equal(userBalance, u.Balance)
	t.Equal(http.StatusOK, resp.StatusCode)
}

func (t *handlerSuite) Test_withdrawBalanceNotEnoughMoney() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	money := -1000.0

	rep := mocks.NewMockRepository()
	rep.On("ChangeBalance", userId, money).Return(nil, postgresdb.ErrorNotEnoughMoney)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"id":    userId,
			"money": money,
		},
	)
	t.Nil(err)

	req, err := http.NewRequest("POST", testSrv.URL+"/changeBalance", bytes.NewReader(body))
	t.Nil(err)

	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)

	t.Contains(string(body), postgresdb.ErrorNotEnoughMoney.Error())
	t.Equal(http.StatusOK, resp.StatusCode)
}

func (t *handlerSuite) Test_addBalanceSomeError() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	money := 50.0

	rep := mocks.NewMockRepository()
	rep.On("ChangeBalance", userId, money).Return(nil, fmt.Errorf("some error"))

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"id":    userId,
			"money": money,
		},
	)
	t.Nil(err)

	req, err := http.NewRequest("POST", testSrv.URL+"/changeBalance", bytes.NewReader(body))
	t.Nil(err)

	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	t.Equal(http.StatusInternalServerError, resp.StatusCode)
}

func (t *handlerSuite) Test_transferBalanceSuccess() {
	fromId := "f0812ab6-9993-11ec-b909-0242ac120002"
	toId := "f0812ab6-9993-11ec-b909-0242ac120003"
	amount := 50.0

	rep := mocks.NewMockRepository()
	rep.On("TransferBalance", fromId, toId, amount).Return(nil)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"to_id":   toId,
			"from_id": fromId,
			"money":   amount,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/transferBalance", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	t.Equal(http.StatusOK, resp.StatusCode)
}

func (t *handlerSuite) Test_transferBalanceNotEnoughMoney() {
	fromId := "f0812ab6-9993-11ec-b909-0242ac120002"
	toId := "f0812ab6-9993-11ec-b909-0242ac120003"
	amount := 50.0

	rep := mocks.NewMockRepository()
	rep.On("TransferBalance", fromId, toId, amount).Return(postgresdb.ErrorNotEnoughMoney)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"to_id":   toId,
			"from_id": fromId,
			"money":   amount,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/transferBalance", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)

	t.Contains(string(body), postgresdb.ErrorNotEnoughMoney.Error())
	t.Equal(http.StatusOK, resp.StatusCode)
}

func (t *handlerSuite) Test_transferBalanceUserNotFound() {
	fromId := "f0812ab6-9993-11ec-b909-0242ac120002"
	toId := "f0812ab6-9993-11ec-b909-0242ac120003"
	amount := 50.0

	rep := mocks.NewMockRepository()
	rep.On("TransferBalance", fromId, toId, amount).Return(postgresdb.ErrorUserNotFound)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"to_id":   toId,
			"from_id": fromId,
			"money":   amount,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/transferBalance", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)

	t.Contains(string(body), postgresdb.ErrorUserNotFound.Error())

	t.Equal(http.StatusNotFound, resp.StatusCode)
}

func (t *handlerSuite) Test_transferSomeError() {
	fromId := "f0812ab6-9993-11ec-b909-0242ac120002"
	toId := "f0812ab6-9993-11ec-b909-0242ac120003"
	amount := 50.0

	rep := mocks.NewMockRepository()
	rep.On("TransferBalance", fromId, toId, amount).Return(fmt.Errorf("some error"))

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"to_id":   toId,
			"from_id": fromId,
			"money":   amount,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/transferBalance", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	t.Equal(http.StatusInternalServerError, resp.StatusCode)
}

func (t *handlerSuite) Test_reserveMoneySuccess() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	serviceId := "f0812ab6-9993-11ec-b909-0242ac120003"
	orderId := "f0812ab6-9993-11ec-b909-0242ac120004"
	amount := 50.0

	rep := mocks.NewMockRepository()
	rep.On("ReserveMoney", userId, serviceId, orderId, amount).Return(nil)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"user_id":    userId,
			"service_id": serviceId,
			"order_id":   orderId,
			"amount":     amount,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/reserveMoney", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	t.Equal(http.StatusOK, resp.StatusCode)
}

func (t *handlerSuite) Test_reserveMoneyNotEnoughMoney() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	serviceId := "f0812ab6-9993-11ec-b909-0242ac120003"
	orderId := "f0812ab6-9993-11ec-b909-0242ac120004"
	amount := 50.0

	rep := mocks.NewMockRepository()
	rep.On("ReserveMoney", userId, serviceId, orderId, amount).Return(postgresdb.ErrorNotEnoughMoney)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"user_id":    userId,
			"service_id": serviceId,
			"order_id":   orderId,
			"amount":     amount,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/reserveMoney", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)

	t.Contains(string(body), postgresdb.ErrorNotEnoughMoney.Error())
	t.Equal(http.StatusOK, resp.StatusCode)
}

func (t *handlerSuite) Test_reserveMoneyUserNotFound() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	serviceId := "f0812ab6-9993-11ec-b909-0242ac120003"
	orderId := "f0812ab6-9993-11ec-b909-0242ac120004"
	amount := 50.0

	rep := mocks.NewMockRepository()
	rep.On("ReserveMoney", userId, serviceId, orderId, amount).Return(postgresdb.ErrorUserNotFound)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"user_id":    userId,
			"service_id": serviceId,
			"order_id":   orderId,
			"amount":     amount,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/reserveMoney", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)

	t.Contains(string(body), postgresdb.ErrorUserNotFound.Error())
	t.Equal(http.StatusNotFound, resp.StatusCode)
}

func (t *handlerSuite) Test_reserveMoneyBadRequest() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	serviceId := "f0812ab6-9993-11ec-b909-0242ac120003"
	orderId := "f0812ab6-9993-11ec-b909-0242ac120004"
	amount := 50.0

	rep := mocks.NewMockRepository()
	rep.On("ReserveMoney", userId, serviceId, orderId, amount).Return(postgresdb.ErrorNegativeAmount)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"user_id":    userId,
			"service_id": serviceId,
			"order_id":   orderId,
			"amount":     amount,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/reserveMoney", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	t.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (t *handlerSuite) Test_deReserveMoneySuccess() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	serviceId := "f0812ab6-9993-11ec-b909-0242ac120003"
	orderId := "f0812ab6-9993-11ec-b909-0242ac120004"
	amount := 50.0

	rep := mocks.NewMockRepository()
	rep.On("DeReserveMoney", userId, serviceId, orderId, amount).Return(nil)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"user_id":    userId,
			"service_id": serviceId,
			"order_id":   orderId,
			"amount":     amount,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/deReserveMoney", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	t.Equal(http.StatusOK, resp.StatusCode)
}

func (t *handlerSuite) Test_deReserveMoneyAlreadyDeReserved() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	serviceId := "f0812ab6-9993-11ec-b909-0242ac120003"
	orderId := "f0812ab6-9993-11ec-b909-0242ac120004"
	amount := 50.0

	rep := mocks.NewMockRepository()
	rep.On("DeReserveMoney", userId, serviceId, orderId, amount).Return(postgresdb.ErrorReserveAlreadyDeReserved)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"user_id":    userId,
			"service_id": serviceId,
			"order_id":   orderId,
			"amount":     amount,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/deReserveMoney", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)

	t.Contains(string(body), postgresdb.ErrorReserveAlreadyDeReserved.Error())
	t.Equal(http.StatusOK, resp.StatusCode)
}

func (t *handlerSuite) Test_RecognizeMoneySuccess() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	serviceId := "f0812ab6-9993-11ec-b909-0242ac120003"
	orderId := "f0812ab6-9993-11ec-b909-0242ac120004"
	amount := 50.0

	rep := mocks.NewMockRepository()
	rep.On("RecognizedMoney", userId, serviceId, orderId, amount).Return(nil)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"user_id":    userId,
			"service_id": serviceId,
			"order_id":   orderId,
			"amount":     amount,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/recognizeMoney", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	t.Equal(http.StatusOK, resp.StatusCode)
}

func (t *handlerSuite) Test_RecognizeMoneyReserveNotFound() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	serviceId := "f0812ab6-9993-11ec-b909-0242ac120003"
	orderId := "f0812ab6-9993-11ec-b909-0242ac120004"
	amount := 50.0

	rep := mocks.NewMockRepository()
	rep.On("RecognizedMoney", userId, serviceId, orderId, amount).Return(postgresdb.ErrorReserveNotFound)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"user_id":    userId,
			"service_id": serviceId,
			"order_id":   orderId,
			"amount":     amount,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/recognizeMoney", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)

	t.Contains(string(body), postgresdb.ErrorReserveNotFound.Error())

	t.Equal(http.StatusNotFound, resp.StatusCode)
}

func (t *handlerSuite) Test_RecognizeMoneyAlreadyRecognized() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	serviceId := "f0812ab6-9993-11ec-b909-0242ac120003"
	orderId := "f0812ab6-9993-11ec-b909-0242ac120004"
	amount := 50.0

	rep := mocks.NewMockRepository()
	rep.On("RecognizedMoney", userId, serviceId, orderId, amount).Return(postgresdb.ErrorReserveAlreadyRecognized)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"user_id":    userId,
			"service_id": serviceId,
			"order_id":   orderId,
			"amount":     amount,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/recognizeMoney", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)

	t.Contains(string(body), postgresdb.ErrorReserveAlreadyRecognized.Error())

	t.Equal(http.StatusOK, resp.StatusCode)
}

func (t *handlerSuite) Test_GetAllTransactions() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	sortType := "date_asc"
	limit := 10
	page := 1

	transaction1 := models.Transaction{
		Id: "some id",
	}

	transaction2 := models.Transaction{
		Id: "some id2",
	}

	returnTransactions := []models.Transaction{transaction1, transaction2}

	rep := mocks.NewMockRepository()
	rep.On("GetAllTransactions", userId, sortType, limit, page).Return(&returnTransactions, nil)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"id":        userId,
			"sort_type": sortType,
			"limit":     limit,
			"page":      page,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/allTransactions", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	var transactions []models.Transaction
	json.NewDecoder(resp.Body).Decode(&transactions)

	t.Equal(http.StatusOK, resp.StatusCode)
	t.Equal(transaction1.Id, transactions[0].Id)
}

func (t *handlerSuite) Test_GetAllTransactionsInvalidSortParameters() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	sortType := "date_asc"
	limit := -10
	page := 1

	rep := mocks.NewMockRepository()
	rep.On("GetAllTransactions", userId, sortType, limit, page).Return(nil, postgresdb.ErrorInvalidSortParameters)

	h := handlers.NewHandler(rep)

	testSrv := httptest.NewServer(h.InitRoutes())
	client := testSrv.Client()

	body, err := json.Marshal(
		map[string]interface{}{
			"id":        userId,
			"sort_type": sortType,
			"limit":     limit,
			"page":      page,
		},
	)
	t.Nil(err)
	req, err := http.NewRequest("POST", testSrv.URL+"/allTransactions", bytes.NewReader(body))
	t.Nil(err)
	resp, err := client.Do(req)
	t.Nil(err)
	defer resp.Body.Close()

	transactions := []models.Transaction{}
	json.NewDecoder(resp.Body).Decode(&transactions)

	t.Equal(http.StatusBadRequest, resp.StatusCode)
}
