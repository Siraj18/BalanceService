package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
	"github.com/siraj18/balance-service-new/internal/db/postgresdb"
	"github.com/siraj18/balance-service-new/internal/handlers"
	"github.com/siraj18/balance-service-new/internal/models"
	"github.com/siraj18/balance-service-new/pkg/postgres"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Required running docker

type PostgreSQLContainer struct {
	testcontainers.Container
	MappedPort string
	Host       string
}

type TestSuite struct {
	suite.Suite
	psqlContainer *PostgreSQLContainer
	server        *httptest.Server
}

func (s *TestSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	psqlContainer, err := NewPostgreSQLContainer(ctx)
	s.Require().NoError(err)
	s.psqlContainer = psqlContainer

	db, err := postgres.NewDb(psqlContainer.GetDSN(), 10)
	if err != nil {
		logrus.Fatal(err)
	}

	rep, err := postgresdb.NewSqlRepository(db)
	if err != nil {
		logrus.Fatal(err)
	}

	handler := handlers.NewHandler(rep)

	s.server = httptest.NewServer(handler.InitRoutes())

}

func (s *TestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.psqlContainer.Terminate(ctx))

	s.server.Close()
}

func NewPostgreSQLContainer(ctx context.Context) (*PostgreSQLContainer, error) {
	req := testcontainers.ContainerRequest{
		Env: map[string]string{
			"POSTGRES_USER":     "user",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "postgres_test",
		},
		ExposedPorts: []string{"5432/tcp"},
		Image:        "postgres:14.2-alpine",
		WaitingFor: wait.ForExec([]string{"pg_isready", "-d", "postgres_test", "-U", "user"}).
			WithPollInterval(1 * time.Second).
			WithExitCodeMatcher(func(exitCode int) bool {
				return exitCode == 0
			}),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	return &PostgreSQLContainer{
		Container:  container,
		MappedPort: mappedPort.Port(),
		Host:       host,
	}, nil
}

func (c PostgreSQLContainer) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", "user", "password", c.Host, c.MappedPort, "postgres_test")
}

func TestSuite_Run(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestChangeBalance() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"
	userBalance := 50.0

	body, err := json.Marshal(
		map[string]interface{}{
			"id":    userId,
			"money": userBalance,
		},
	)
	s.Nil(err)

	res, err := s.server.Client().Post(s.server.URL+"/changeBalance", "", bytes.NewReader(body))
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	response := models.User{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Assert().Equal(userId, response.Id)
	s.Assert().Equal(userBalance, response.Balance)
}

func (s *TestSuite) TestGetBalance() {
	userId := "f0812ab6-9993-11ec-b909-0242ac120002"

	res, err := s.server.Client().Get(s.server.URL + "/balance/" + userId)
	s.Require().NoError(err)

	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	response := models.User{}
	err = json.NewDecoder(res.Body).Decode(&response)
	s.Require().NoError(err)

	s.Assert().Equal(userId, response.Id)
}
