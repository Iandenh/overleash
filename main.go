package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/Iandenh/overleash/config"
	"github.com/Iandenh/overleash/overleash"
	"github.com/Iandenh/overleash/server"
	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

func run(ctx context.Context, cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	o := overleash.NewOverleash(cfg)
	o.Start(ctx)

	server.New(o, ctx).Start()
}

func main() {
	log.Info("Starting Overleash")
	cfg, err := config.InitConfig()

	if err != nil {
		log.Fatal(err)
	}

	if cfg.Verbose {
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug logging enabled")
		log.Debug(viper.AllSettings())
	}

	run(context.Background(), cfg)
}
