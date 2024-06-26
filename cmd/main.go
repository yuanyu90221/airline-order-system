package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yuanyu90221/airline-order-system/internal/application"
	"github.com/yuanyu90221/airline-order-system/internal/config"
)

func main() {
	app := application.New(config.AppConfig)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	err := app.Start(ctx)
	if err != nil {
		log.Println("failed to start app:", err)
	}
}
