package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backoffice-service/config"
	deliveryHTTP "backoffice-service/internal/delivery/http"
	"backoffice-service/internal/pkg/database"
	"backoffice-service/internal/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	// 1. Load Configurations
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// 2. Initialize Structured Logger
	log, err := logger.NewLogger(cfg.App.Env)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Starting Loyalty Platform BackOffice Service",
		zap.String("env", cfg.App.Env),
		zap.String("port", cfg.App.Port),
	)

	// 3. Graceful Shutdown Notification Context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 4. Initialize Database Connection Pool
	db, err := database.NewPostgresClient(ctx, cfg.DB, log)
	if err != nil {
		log.Fatal("Failed to initialize database client", zap.Error(err))
	}
	defer db.Close()

	// 5. Initialize HTTP Router
	router := deliveryHTTP.NewRouter(cfg, db, log)

	// 6. Start HTTP Server
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("HTTP server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server failed to listen", zap.Error(err))
		}
	}()

	// 7. Wait for OS Interrupt / Termination Signal
	<-ctx.Done()
	log.Info("Shutting down BackOffice HTTP server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("BackOffice service stopped cleanly")
}
