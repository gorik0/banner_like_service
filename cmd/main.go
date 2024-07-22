package main

import (
	. "baner_service/cmd/error"
	"baner_service/internal/cache"
	"baner_service/internal/config"
	"baner_service/internal/db"
	"context"
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

	//	::: SERVER setup

	//::: server shutdown

	//::: server run

}
