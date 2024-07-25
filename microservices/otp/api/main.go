package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"midaslabs/microservices/otp/internal/application"
	"midaslabs/microservices/otp/internal/infrastructure"
	"midaslabs/sdk/rabbitmq"

	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/arl/statsviz"
	"github.com/charmbracelet/log"
)

var build = "develop"

func main() {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		Prefix:          "OTP",
		ReportTimestamp: true,
	})

	ctx := context.Background()

	if err := run(ctx, logger); err != nil {
		logger.Error(ctx, "startup", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *log.Logger) error {
	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:3010"`
		}
		Twilio struct {
			AccountSID string `conf:"mask"`
			AuthToken  string `conf:"mask"`
			ServiceSID string `conf:"mask"`
		}
		RabbitMQ struct {
			URL string `conf:"default:amqp://guest:guest@localhost:5672/"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "OTP Service",
		},
	}

	const prefix = "OTP"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	logger.Info("starting service", "version", cfg.Build)
	defer logger.Info("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	logger.Info("startup", "config", out)

	expvar.NewString("build").Set(cfg.Build)

	go func() {
		logger.Info("startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

		if err := http.ListenAndServe(cfg.Web.DebugHost, DebugMux()); err != nil {
			logger.Error("shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	logger.Info("startup", "status", "initializing V1 API support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	messageBroker, err := rabbitmq.NewRabbitMQBroker(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ message broker: %v", err)
	}
	defer messageBroker.Close()

	otpServiceClient := infrastructure.NewMockOTPService(cfg.Twilio.AccountSID, cfg.Twilio.AuthToken, cfg.Twilio.ServiceSID)
	otpService := application.NewOTPService(messageBroker, otpServiceClient)

	go otpService.Start(ctx)
	mux := http.NewServeMux()
	server := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      mux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("startup", "status", "api router started", "host", cfg.Web.APIHost)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		logger.Info("shutdown", "status", "shutdown started", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			server.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

func DebugMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/vars", expvar.Handler().ServeHTTP)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Register statsviz handlers on the mux.
	statsviz.Register(mux)

	return mux
}
