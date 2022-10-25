package main

import (
	"github.com/siraj18/balance-service-new/internal/db/postgresdb"
	"github.com/siraj18/balance-service-new/internal/handlers"
	"github.com/siraj18/balance-service-new/internal/server"
	"github.com/siraj18/balance-service-new/pkg/postgres"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

// @title Balance Service API
// @version 1.0
// @description api for balance service
// @BasePath /
func main() {
	conStr := os.Getenv("connection_string_postgres")
	address := os.Getenv("address")

	db, err := postgres.NewDb(conStr, 10)
	if err != nil {
		logrus.Fatal(err)
	}

	rep, err := postgresdb.NewSqlRepository(db)
	if err != nil {
		logrus.Fatal(err)
	}

	handler := handlers.NewHandler(rep)

	server := server.NewServer(address, handler.InitRoutes(), time.Second*10)
	if err := server.Run(); err != nil {
		logrus.Fatal(err)
	}
}
