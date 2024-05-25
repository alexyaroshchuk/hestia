package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hestia/pkg"
	"hestia/pkg/auth"
	"hestia/pkg/db"
	"hestia/pkg/middlewares"
	"hestia/pkg/services"
	"hestia/pkg/web"

	_ "github.com/lib/pq"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	os.Exit(run(ctx, os.Stderr))
}

func run(ctx context.Context, w io.Writer) int {
	logger := slog.New(slog.NewTextHandler(w, nil))
	logger = logger.With("revision", pkg.BuildRevision, "time", pkg.BuildRevisionTime)

	cfg, err := configFromEnv()
	if err != nil {
		logger.Error("failed to get config from environment", "error", err)
		return 1
	}

	dbPG, err := connectPGSQL(cfg)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		return 1
	}

	defer func() {
		err := dbPG.Close()
		if err != nil {
			logger.Error("failed to close database handles", "error", err)
			return
		}
	}()

	usrErrHandler := func(err error) {
		logger.Error("user service error", "error", err)
	}
	userSvc := services.NewUserService(dbPG, usrErrHandler)

	jwtC := auth.NewServiceConfig(cfg.auth.SecretKey, cfg.auth.TokenDuration)
	interceptor := auth.NewAuthInterceptor(jwtC, middlewares.AccessibleRoles())

	emailErrHandler := func(err error) {
		logger.Error("email service error", "error", err)
	}
	emailSvc := services.NewEmailService(dbPG, emailErrHandler)

	authErrHandler := func(err error) {
		logger.Error("auth service error", "error", err)
	}
	authSvc := services.NewAuthServer(dbPG, jwtC, authErrHandler)

	flatErrHandler := func(err error) {
		logger.Error("flat service error", "error", err)
	}
	flatSvc := services.NewFlatService(dbPG, flatErrHandler)

	serverDeps := &web.ServerDeps{
		Logger:       logger,
		AuthService:  authSvc,
		UserService:  userSvc,
		FlatService:  flatSvc,
		EmailService: emailSvc,
		JWT:          jwtC,
		Interceptor:  interceptor,
	}

	srv := &http.Server{
		Addr:         cfg.http.addr,
		ReadTimeout:  cfg.http.readTimeout,
		WriteTimeout: cfg.http.writeTimeout,
		IdleTimeout:  cfg.http.idleTimeout,
		Handler:      web.NewServer(serverDeps),
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.Info("starting http server", "addr", cfg.http.addr)
		return srv.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()
		logger.Info("stopping http server")

		shutCtx, cancel := context.WithTimeout(context.Background(), cfg.http.shutdownTimeout)
		defer cancel()

		return srv.Shutdown(shutCtx)
	})

	err = g.Wait()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("http server stopped with error", "error", err)
		return 1
	}

	logger.Info("http server stopped successfully")

	return 0
}

// connectPGSQL connects to the database.
func connectPGSQL(cfg config) (*sql.DB, error) {
	dbPG, err := db.OpenPGSQL(cfg.db.connection)
	if err != nil {
		closeErr := dbPG.Close()
		if closeErr != nil {
			err = errors.Join(err, closeErr)
		}
		return nil, fmt.Errorf("failed to open database handle: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = dbPG.PingContext(ctx)
	if err != nil {
		closeErr := dbPG.Close()
		if closeErr != nil {
			err = errors.Join(err, closeErr)
		}
		return nil, fmt.Errorf("failed to ping: %w", err)
	}

	return dbPG, nil
}
