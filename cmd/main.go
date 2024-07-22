package main

import (
	. "baner_service/cmd/error"
	"baner_service/internal/config"
)

func main() {

	//	:::: load CONFIG
	cfg, err := config.LoadConfig("config.yaml")
	HandleErr(err, "load config")

	//	::: CACHE setup

	//	::: DB setup

	//	::: SERVER setup

	//::: server shutdown

	//::: server run

}
