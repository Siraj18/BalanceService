package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	_ "github.com/siraj18/balance-service-new/docs"
	"github.com/siraj18/balance-service-new/internal/db/postgresdb"
	"github.com/siraj18/balance-service-new/internal/models"
	"github.com/siraj18/balance-service-new/internal/utils"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
	"io"
	"net/http"
	"os"
)

type handler struct {
	router     *chi.Mux
	logger     *logrus.Logger
	repository Repository
}

func NewHandler(rep Repository) *handler {
	return &handler{
		router:     chi.NewRouter(),
		logger:     logrus.New(),
		repository: rep,
	}
}

type Repository interface {
	GetBalance(string) (*models.User, error)
	ChangeBalance(string, float64) (*models.User, error)
	TransferBalance(string, string, float64) error
	GetAllTransactions(string, string, int, int) (*[]models.Transaction, error)
	ReserveMoney(string, string, string, float64) error
	RecognizedMoney(string, string, string, float64) error
	DeReserveMoney(string, string, string, float64) error
	GetReserves(int, int) (*[]models.Reserve, error)
}

// handler - Returns all the available APIs
// GetBalance godoc
// @Summary      Get account balance
// @Description  get balance by UID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param   uid   path    string  true  "User account ID" default(34be95d0-9a41-11ec-b909-0242ac120003)
// @Success 200 {object} models.User
// @Failure      400  {string} string
// @Failure      404  {string} string
// @Router /balance/{uid} [get]
func (handler *handler) getBalance(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "uid")

	user, err := handler.repository.GetBalance(uid)

	if err != nil {
		if errors.Is(err, postgresdb.ErrorUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, postgresdb.ErrorInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(500)
		handler.logger.Error(err)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// ChangeBalance godoc
// @Summary      Change user account balance or create account
// @Description  change user account balance by uid or create account
// @Tags         users
// @Accept       json
// @Produce      json
// @Param   account   body    models.UserChangeBalanceQuery  true  "Account"
// @Success 200 {object} models.User
// @Failure      400  {string} string
// @Router /changeBalance [post]
func (handler *handler) changeBalance(w http.ResponseWriter, r *http.Request) {
	var postData models.UserChangeBalanceQuery

	err := json.NewDecoder(r.Body).Decode(&postData)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "invalid post data", http.StatusBadRequest)
		return
	}

	user, err := handler.repository.ChangeBalance(postData.Id, postData.Money)
	if err != nil {
		if errors.Is(err, postgresdb.ErrorNotEnoughMoney) {
			http.Error(w, err.Error(), http.StatusOK)
			return
		}

		if errors.Is(err, postgresdb.ErrorInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		handler.logger.Error(err)
		return
	}
	json.NewEncoder(w).Encode(user)
}

// TransferBalance godoc
// @Summary      Transferring money from one user account to another
// @Description  transferring money from one user account to another
// @Tags         users
// @Accept       json
// @Produce      json
// @Param   account   body    models.UserTransferBalanceQuery  true  "Account"
// @Success 200 {string} string
// @Failure      400  {string} string
// @Router /transferBalance [post]
func (handler *handler) transferBalance(w http.ResponseWriter, r *http.Request) {
	var postData models.UserTransferBalanceQuery

	err := json.NewDecoder(r.Body).Decode(&postData)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "invalid post data", http.StatusBadRequest)
		return
	}

	err = handler.repository.TransferBalance(postData.FromId, postData.ToId, postData.Money)

	if err != nil {
		switch err {
		case postgresdb.ErrorNotEnoughMoney:
			http.Error(w, err.Error(), http.StatusOK)
		case postgresdb.ErrorUserNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case postgresdb.ErrorInvalidInput, postgresdb.ErrorNegativeAmount:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		handler.logger.Error(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "The transfer was completed successfully")
}

// GetAllTransactions godoc
// @Summary      Get all transactions
// @Description  Get all transactions by uuid
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param   account   body    models.AllTransactionsGetQuery  true  "TransactionParams"
// @Success 200 {object} []models.Transaction
// @Failure      400  {object} string
// @Router /allTransactions [post]
func (handler *handler) getAllTransactions(w http.ResponseWriter, r *http.Request) {
	var postData models.AllTransactionsGetQuery

	err := json.NewDecoder(r.Body).Decode(&postData)
	defer r.Body.Close()

	if err != nil {
		handler.logger.Error(err)
		http.Error(w, "invalid post data", http.StatusBadRequest)
		return
	}

	transactions, err := handler.repository.GetAllTransactions(postData.Id, postData.SortType, postData.Limit, postData.Page)
	if err != nil {
		if errors.Is(err, postgresdb.ErrorInvalidSortParameters) || errors.Is(err, postgresdb.ErrorInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		handler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(transactions)

}

// ReserveMoney godoc
// @Summary      reserving money from the user account
// @Description  reserving money from the user account
// @Tags         users
// @Accept       json
// @Produce      json
// @Param   account   body    models.ReserveMoneyQuery  true  "Account"
// @Success 200 {string} string
// @Failure      404  {string} string
// @Router /reserveMoney [post]
func (handler *handler) reserveMoney(w http.ResponseWriter, r *http.Request) {
	var postData models.ReserveMoneyQuery

	err := json.NewDecoder(r.Body).Decode(&postData)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "invalid post data", http.StatusBadRequest)
		return
	}

	err = handler.repository.ReserveMoney(postData.UserId, postData.ServiceId, postData.OrderId, postData.Amount)
	if err != nil {
		if errors.Is(err, postgresdb.ErrorNotEnoughMoney) {
			http.Error(w, err.Error(), http.StatusOK)
			return
		}

		if errors.Is(err, postgresdb.ErrorUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, postgresdb.ErrorNegativeAmount) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, postgresdb.ErrorInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		handler.logger.Error(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "successfully reserved")
}

// DeReserveMoney godoc
// @Summary      de-reserving money from the user account
// @Description  de-reserving money from the user account
// @Tags         users
// @Accept       json
// @Produce      json
// @Param   account   body    models.ReserveMoneyQuery  true  "Account"
// @Success 200 {string} string
// @Failure      404  {string} string
// @Router /deReserveMoney [post]
func (handler *handler) deReserveMoney(w http.ResponseWriter, r *http.Request) {
	var postData models.ReserveMoneyQuery

	err := json.NewDecoder(r.Body).Decode(&postData)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "invalid post data", http.StatusBadRequest)
		return
	}

	err = handler.repository.DeReserveMoney(postData.UserId, postData.ServiceId, postData.OrderId, postData.Amount)
	if err != nil {
		if errors.Is(err, postgresdb.ErrorReserveAlreadyRecognized) {
			http.Error(w, err.Error(), http.StatusOK)
			return
		}

		if errors.Is(err, postgresdb.ErrorReserveAlreadyDeReserved) {
			http.Error(w, err.Error(), http.StatusOK)
			return
		}

		if errors.Is(err, postgresdb.ErrorReserveNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, postgresdb.ErrorInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		handler.logger.Error(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "successfully de-reserved")
}

// RecognizeMoney godoc
// @Summary      recognize money from the reserve account
// @Description  recognize money from the reserve account
// @Tags         users
// @Accept       json
// @Produce      json
// @Param   account   body    models.ReserveMoneyQuery  true  "Account"
// @Success 200 {string} string
// @Failure      404  {string} string
// @Router /recognizeMoney [post]
func (handler *handler) recognizeMoney(w http.ResponseWriter, r *http.Request) {
	var postData models.ReserveMoneyQuery

	err := json.NewDecoder(r.Body).Decode(&postData)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "invalid post data", http.StatusBadRequest)
		return
	}

	err = handler.repository.RecognizedMoney(postData.UserId, postData.ServiceId, postData.OrderId, postData.Amount)
	if err != nil {
		switch err {
		case postgresdb.ErrorReserveNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case postgresdb.ErrorUserNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case postgresdb.ErrorInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case postgresdb.ErrorReserveAlreadyRecognized, postgresdb.ErrorReserveAlreadyDeReserved:
			http.Error(w, err.Error(), http.StatusOK)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		handler.logger.Error(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "successfully recognized")
}

// GetReportLink godoc
// @Summary      Get all transactions
// @Description  Get all transactions by uuid
// @Tags         reports
// @Accept       json
// @Produce      json
// @Param   account   body    models.GetReportLinkQuery  true  "ReportsParams"
// @Success 200 {string} string
// @Failure      400  {object} string
// @Router /getReportLink [post]
func (handler *handler) getReportLink(w http.ResponseWriter, r *http.Request) {
	var postData models.GetReportLinkQuery

	err := json.NewDecoder(r.Body).Decode(&postData)
	defer r.Body.Close()

	if err != nil {
		handler.logger.Error(err)
		http.Error(w, "invalid post data", http.StatusBadRequest)
		return
	}

	reserves, err := handler.repository.GetReserves(postData.Year, postData.Month)
	if err != nil {
		handler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	link, err := utils.GenerateReportsLink(reserves, r.Host)
	if err != nil {
		handler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, link)
}

func (handler *handler) HandleFile(w http.ResponseWriter, r *http.Request) {
	fileId := chi.URLParam(r, "fileId")

	file, err := os.Open("./files/reports/" + fileId + ".csv")
	defer file.Close()

	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/csv")

	io.Copy(w, file)
}

func (handler *handler) InitRoutes() *chi.Mux {

	handler.router.Get("/balance/{uid}", handler.getBalance)
	handler.router.Post("/changeBalance", handler.changeBalance)
	handler.router.Post("/transferBalance", handler.transferBalance)
	handler.router.Post("/reserveMoney", handler.reserveMoney)
	handler.router.Post("/recognizeMoney", handler.recognizeMoney)
	handler.router.Post("/deReserveMoney", handler.deReserveMoney)

	handler.router.Post("/allTransactions", handler.getAllTransactions)

	handler.router.Get("/reports/{fileId}", handler.HandleFile)
	handler.router.Post("/getReportLink", handler.getReportLink)

	handler.router.Mount("/swagger", httpSwagger.WrapHandler)

	return handler.router
}
