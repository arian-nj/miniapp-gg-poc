package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/go-telegram/miniapp/internal/application"
	"github.com/go-telegram/miniapp/internal/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg := config.Config{
		ListenAddress: "0.0.0.0:3000",
	}

	if v := os.Getenv("LISTEN_ADDRESS"); v != "" {
		cfg.ListenAddress = v
	}
	if v := os.Getenv("TG_BOT_TOKEN"); v != "" {
		cfg.TgBotToken = v
	}

	err := run(ctx, cancel, cfg)
	if err != nil {
		log.Printf("error run: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cancel context.CancelFunc, cfg config.Config) error {
	ln, errLn := net.Listen("tcp", cfg.ListenAddress)
	if errLn != nil {
		return fmt.Errorf("failed to listen: %w", errLn)
	}
	defer ln.Close()

	app, errApp := application.New(cfg.TgBotToken)
	if errApp != nil {
		return fmt.Errorf("failed to create application: %w", errApp)
	}

	log.Printf("start")

	var wg sync.WaitGroup

	wg.Add(1)
	go app.Run(ctx, cancel, &wg, ln)

	<-ctx.Done()

	wg.Wait()

	log.Printf("done")

	return nil
}
