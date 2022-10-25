package server

import (
	"context"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type server struct {
	server *http.Server
	logger *logrus.Logger
}

func NewServer(addr string, handler http.Handler, timeouts time.Duration) *server {
	return &server{
		server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  timeouts,
			WriteTimeout: timeouts,
		},
		logger: logrus.New(),
	}
}

func (s *server) Run() error {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-exit
		s.Stop()

	}()

	s.logger.Info("starting server on port" + s.server.Addr)
	err := s.server.ListenAndServe()

	if err == http.ErrServerClosed {
		return nil
	}

	return err
}

func (s *server) Stop() {
	s.logger.Info("stopping server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.server.Shutdown(ctx)

}
