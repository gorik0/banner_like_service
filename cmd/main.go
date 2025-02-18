package main

import (
	. "baner_service/cmd/error"
	"baner_service/cmd/router"
	"baner_service/internal/cache"
	"baner_service/internal/config"
	"baner_service/internal/db"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	//::: CONTEXT
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	//	:::: load CONFIG
	cfg, err := config.LoadConfig("config.yaml")
	HandleErr(err, "load config")

	//	::: CACHE setup
	redis, err := cache.NewRedis(cfg.Redis)
	HandleErr(err, "init redis")

	//	::: DB setup
	connDbCtx, cancelFunc := context.WithTimeout(ctx, cfg.Cancel*time.Second)
	defer cancelFunc()
	db, err := db.NewPostgres(connDbCtx, cfg.DBConnect)
	HandleErr(err, "init db")
	//	::: SERVER setup

	server := http.Server{
		Addr:    fmt.Sprintf(":" + cfg.Port),
		Handler: router.Register(redis, db, cfg.UserToken, cfg.AdminToken),
	}
	//::: server shutdown

	go func() {
		<-ctx.Done()
		ctxShutdown, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		server.Shutdown(ctxShutdown)

	}()
	//::: server run

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		HandleErr(err, "server")
	}
}
