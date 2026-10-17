package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"inventory/internal/config"
	"inventory/internal/handler"
	"inventory/internal/middleware"
	"inventory/internal/repository"
	"inventory/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database config: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("database unreachable: %v", err)
	}
	if err := migrate(ctx, pool); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	productSvc := service.NewProductService(repository.NewProductRepository(pool), cfg.LowStockThreshold)
	categorySvc := service.NewCategoryService(repository.NewCategoryRepository(pool))
	router := handler.Routes(handler.NewProductHandler(productSvc), handler.NewCategoryHandler(categorySvc))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      middleware.Logger(middleware.CORS(cfg.AllowedOrigin)(router)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("API listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
