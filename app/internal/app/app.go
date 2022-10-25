package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/slava-911/URL-shortener/internal/adapter/db"
	"github.com/slava-911/URL-shortener/internal/config"
	"github.com/slava-911/URL-shortener/internal/controller/http/handler"
	"github.com/slava-911/URL-shortener/internal/domain/service"
	"github.com/slava-911/URL-shortener/internal/jwt"
	"github.com/slava-911/URL-shortener/pkg/cache/freecache"
	"github.com/slava-911/URL-shortener/pkg/logging"
	"github.com/slava-911/URL-shortener/pkg/metric"
	"github.com/slava-911/URL-shortener/pkg/postgresql"
)

type App struct {
	cfg        *config.Config
	logger     *logging.Logger
	router     *httprouter.Router
	httpServer *http.Server
	dbClient   postgresql.Client
}

func NewApp(config *config.Config, logger *logging.Logger) (App, error) {
	logger.Info("router initialization")
	router := httprouter.New()
	router.Handler(http.MethodGet, "/", http.RedirectHandler("/links", http.StatusMovedPermanently))

	//logger.Info("swagger docs initialization")
	//router.Handler(http.MethodGet, "/swagger", http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently))
	//router.Handler(http.MethodGet, "/swagger/*any", httpSwagger.WrapHandler)

	dbConfig := postgresql.NewDBConfig(
		config.PostgreSQL.Username, config.PostgreSQL.Password,
		config.PostgreSQL.Host, config.PostgreSQL.Port, config.PostgreSQL.Database,
	)
	dbClient, err := postgresql.NewClient(5, 5*time.Second, dbConfig, logger)
	if err != nil {
		logger.Fatal(err)
	}

	validate := validator.New()

	logger.Println("cache initialization")
	refreshTokenCache := freecache.NewCacheRepo(104857600) // 100MB

	logger.Println("helpers initialization")
	jwtHelper := jwt.NewHelper(refreshTokenCache, logger)

	logger.Info("create and register handlers")

	logger.Info("heartbeat metric initialization")
	metricHandler := metric.Handler{}
	metricHandler.Register(router)

	userStorage := db.NewUserStorage(dbClient, logger)
	userService := service.NewUserService(userStorage, logger)
	userHandler := handler.NewUserHandler(jwtHelper, userService, validate, logger)
	userHandler.Register(router)

	linkStorage := db.NewLinkStorage(dbClient, logger)
	linkService := service.NewLinkService(linkStorage, logger)
	linkHandler := handler.NewLinkHandler(linkService, validate, logger)
	linkHandler.Register(router)

	return App{
		cfg:      config,
		logger:   logger,
		router:   router,
		dbClient: dbClient,
	}, nil
}

func (a *App) Run() {
	a.startHTTP()
}

func (a *App) startHTTP() {
	a.logger.WithFields(map[string]interface{}{
		"IP":   a.cfg.HTTP.IP,
		"Port": a.cfg.HTTP.Port,
	}).Info("HTTP Server initializing")

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.cfg.HTTP.IP, a.cfg.HTTP.Port))
	if err != nil {
		a.logger.WithError(err).Fatal("failed to create listener")
	}

	c := cors.New(cors.Options{
		AllowedMethods:     a.cfg.HTTP.CORS.AllowedMethods,
		AllowedOrigins:     a.cfg.HTTP.CORS.AllowedOrigins,
		AllowCredentials:   a.cfg.HTTP.CORS.AllowCredentials,
		AllowedHeaders:     a.cfg.HTTP.CORS.AllowedHeaders,
		OptionsPassthrough: a.cfg.HTTP.CORS.OptionsPassthrough,
		ExposedHeaders:     a.cfg.HTTP.CORS.ExposedHeaders,
		Debug:              a.cfg.HTTP.CORS.Debug,
	})

	cHandler := c.Handler(a.router)

	a.httpServer = &http.Server{
		Handler:      cHandler,
		WriteTimeout: a.cfg.HTTP.WriteTimeout,
		ReadTimeout:  a.cfg.HTTP.ReadTimeout,
	}

	a.logger.Info("application completely initialized and started")

	go doGracefulShutdown(a)

	if err := a.httpServer.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			a.logger.Warn("server shutdown")
		default:
			a.logger.Fatal(err)
		}
	}
}

func doGracefulShutdown(a *App) {

	signals := []os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM}
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, signals...)
	sig := <-sigch
	a.logger.Infof("Caught signal %s. Shutting down...", sig)

	defer a.dbClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := a.httpServer.Shutdown(ctx)
	if err != nil {
		a.logger.Fatal(err)
	}
}
