package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"midaslabs/gen/auth/v1/authv1connect"

	"midaslabs/microservices/auth/api/handlers"
	"midaslabs/microservices/auth/internal/application"
	infrastructure "midaslabs/microservices/auth/internal/infrastrucutre"
	"midaslabs/sdk/rabbitmq"

	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/arl/statsviz"
	"github.com/jmoiron/sqlx"

	"github.com/charmbracelet/log"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var build = "develop"

func main() {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		Prefix:          "AUTH",
		ReportTimestamp: true,
	})

	// -------------------------------------------------------------------------

	ctx := context.Background()

	if err := run(ctx, logger); err != nil {
		logger.Error(ctx, "startup", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *log.Logger) error {

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
			APIHost         string        `conf:"default:0.0.0.0:4000"`
			DebugHost       string        `conf:"default:0.0.0.0:4010"`
			GrpcHost        string        `conf:"default:0.0.0.0:5000"`
		}

		DB struct {
			User         string `conf:"default:user"`
			Password     string `conf:"default:password,mask"`
			Host         string `conf:"default:localhost"`
			Name         string `conf:"default:usersdb"`
			MaxIdleConns int    `conf:"default:0"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}

		OTPProvider struct {
			Host string `conf:"default:http://0.0.0.0:3000"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "AUTH Service",
		},
	}

	const prefix = "AUTH"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// Database Support

	log.Info(ctx, "startup", "status", "initializing database support", "hostport", cfg.DB.Host)

	db, err := OpenDB(DBConfig{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}

	defer db.Close()

	// -------------------------------------------------------------------------
	// App Starting

	logger.Info("starting service", "version", cfg.Build)
	defer logger.Info("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	logger.Info("startup", "config", out)

	expvar.NewString("build").Set(cfg.Build)

	// -------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		logger.Info("startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

		if err := http.ListenAndServe(cfg.Web.DebugHost, DebugMux()); err != nil {
			logger.Error("shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Start API Service

	logger.Info("startup", "status", "initializing V1 API support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	stdlog := logger.StandardLog(log.StandardLogOptions{
		ForceLevel: log.ErrorLevel,
	})

	userRepo := infrastructure.NewPostgresUserRepository(db)
	activityRepo := infrastructure.NewPostgresActivityRepository(db)
	otpRepo := infrastructure.NewPostgresOTPRepository(db)

	// otpClient := infrastructure.NewOTPServiceClient(cfg.OTPProvider.Host)
	rabbitMQURL := "amqp://guest:guest@localhost:5672/"
	messageBroker, err := rabbitmq.NewRabbitMQBroker(rabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ message broker: %v", err)
	}
	defer messageBroker.Close()

	authService := application.NewAuthService(userRepo, activityRepo, otpRepo, messageBroker)
	authHandler := handlers.NewAuthHandler(logger, authService)

	// Start GRPC Server
	go func() {
		logger.Info("startup", "status", "gRPC server started", "host", cfg.Web.GrpcHost)

		if err := http.ListenAndServe(cfg.Web.GrpcHost, GrpcMux(logger, authService)); err != nil {
			logger.Error("shutdown", "status", "Grpc v1 router closed", "host", cfg.Web.GrpcHost, "msg", err)
		}
	}()

	// Start REST API Server
	mux := http.NewServeMux()
	mux.HandleFunc("/signup", authHandler.SignUpWithPhoneNumber)
	mux.HandleFunc("/signup/verify", authHandler.VerifyPhoneNumber)
	mux.HandleFunc("/login/initiate", authHandler.LoginInitiate)
	mux.HandleFunc("/login/complete", authHandler.ValidatePhoneNumberLogin)
	mux.HandleFunc("/profile", authHandler.GetProfile)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      mux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     stdlog,
	}

	serverErrors := make(chan error, 1)

	go func() {

		logger.Info("startup", "status", "api router started", "host", api.Addr)

		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Info("shutdown", "status", "shutdown started", "signal", sig)
		defer logger.Info("shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}
	return nil
}

func DebugMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars/", expvar.Handler())

	statsviz.Register(mux)

	return mux
}

func GrpcMux(logger *log.Logger, authService *application.AuthService) *http.ServeMux {
	mux := http.NewServeMux()
	authServer := handlers.NewAuthServerHandlers(logger, authService)
	path, handler := authv1connect.NewAuthServiceHandler(authServer)
	mux.Handle(path, handler)

	return mux
}

type DBConfig struct {
	User         string
	Password     string
	Host         string
	Name         string
	Schema       string
	MaxIdleConns int
	MaxOpenConns int
	DisableTLS   bool
}

func OpenDB(cfg DBConfig) (*sqlx.DB, error) {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")
	if cfg.Schema != "" {
		q.Set("search_path", cfg.Schema)
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	db, err := sqlx.Open("pgx", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	return db, nil
}
