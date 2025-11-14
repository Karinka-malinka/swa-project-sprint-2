package router

import (
	"net/http"
	"net/url"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/swa-project-sprint-2/src/microservices/proxy/internal/controller/handlers"
	"github.com/swa-project-sprint-2/src/microservices/proxy/internal/controller/server"
	"github.com/swa-project-sprint-2/src/microservices/proxy/internal/lib/config"
)

type Router struct {
	Echo *echo.Echo
	cfg  config.Config
}

func NewRouter(handlers []handlers.Handler, cfg config.Config) *Router {

	router := &Router{Echo: echo.New(), cfg: cfg}

	router.Echo.Validator = &server.CustomValidator{Validator: validator.New()}

	router.Echo.Use(middleware.Logger())

	if cfg.GradualMigration {
		moviesProxy(router.Echo, cfg.MoviesServiceURL)
		eventsProxy(router.Echo, cfg.EventsServiceURL)
	}

	monolithProxy(router.Echo, cfg.MonolithURL)

	router.Echo.GET("/health", handleHealth)

	return router
}

func handleHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]bool{"status": true})
}

func monolithProxy(router *echo.Echo, monolithURL string) {

	catchUrl, err := url.Parse(monolithURL)
	if err != nil {
		logrus.Fatalf("error in parse url. error: %v", err)
	}

	router.Group("/api", middleware.Proxy(
		middleware.NewRandomBalancer([]*middleware.ProxyTarget{
			{
				URL: catchUrl,
			},
		})))

}

func moviesProxy(router *echo.Echo, moviesServiceURL string) {

	catchUrl, err := url.Parse(moviesServiceURL)
	if err != nil {
		logrus.Fatalf("error in parse url. error: %v", err)
	}

	router.Group("/api/movies", middleware.Proxy(
		middleware.NewRandomBalancer([]*middleware.ProxyTarget{
			{
				URL: catchUrl,
			},
		})))

}

func eventsProxy(router *echo.Echo, eventsServiceURL string) {

	catchUrl, err := url.Parse(eventsServiceURL)
	if err != nil {
		logrus.Fatalf("error in parse url. error: %v", err)
	}

	router.Group("/api/events", middleware.Proxy(
		middleware.NewRandomBalancer([]*middleware.ProxyTarget{
			{
				URL: catchUrl,
			},
		})))

}
