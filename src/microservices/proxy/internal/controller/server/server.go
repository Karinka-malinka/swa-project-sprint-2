package server

import (
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/swa-project-sprint-2/src/microservices/proxy/internal/lib/config"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	HTTPServer http.Server
	Config     *config.Config
}

type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.Validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func NewHTTPServer(address string, h http.Handler) *Server {

	server := &Server{}

	server.HTTPServer = http.Server{
		Addr:         address,
		Handler:      h,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	return server

}

func (s *Server) Start(ctx context.Context) {

	errs, _ := errgroup.WithContext(ctx)

	errs.Go(s.HTTPServer.ListenAndServe)

	err := errs.Wait()
	if err != nil {
		logrus.Infof("message from server: %v", err)
	}
}

func (s *Server) Stop(ctx context.Context) {
	err := s.HTTPServer.Shutdown(ctx)
	if err != nil {
		logrus.Errorf("server shutdown with error: %v", err)
	}
	logrus.Infof("Server is graceful shutdown...")
}
