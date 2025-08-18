package main

import (
	"context"
	"os"
	"os/signal"
	"strings"

	"github.com/Iandenh/overleash/overleash"
	"github.com/Iandenh/overleash/server"
	"github.com/charmbracelet/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

import _ "go.uber.org/automaxprocs"

func run(ctx context.Context) {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	tokens := strings.Split(viper.GetString("token"), ",")
	reload := viper.GetInt("reload")
	port := viper.GetInt("port")
	proxyMetrics := viper.GetBool("proxy_metrics")

	upstream := viper.GetString("upstream")
	if upstream == "" {
		upstream = viper.GetString("url")
		if upstream != "" {
			log.Warn("The 'url' flag is deprecated. Please use the 'upstream' flag instead.")
		}
	}

	o := overleash.NewOverleash(upstream, tokens, reload)
	o.Start(ctx)

	server.New(o, port, proxyMetrics, ctx).Start()
}

func main() {
	log.Info("Starting Overleash")
	initConfig()

	if viper.GetBool("verbose") {
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug logging enabled")
		log.Debug(viper.AllSettings())
	}

	run(context.Background())
}

func initConfig() {
	viper.SetEnvPrefix("OVERLEASH")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	pflag.String("url", "", "DEPRECATED: Unleash URL (e.g. https://unleash.my-site.com) without /api. Use --upstream instead.")
	pflag.String("upstream", "", "Unleash upstream URL to load feature flags (e.g. https://unleash.my-site.com) without /api, can be an Unleash instance or Unleash Edge.")
	pflag.String("token", "", "Comma-separated Unleash client token(s) to fetch feature flag configurations.")
	pflag.String("port", "5433", "Port number on which Overleash will listen (default: 5433).")
	pflag.Int("reload", 0, "Reload frequency in minutes for refreshing feature flag configuration (0 disables automatic reloading).")
	pflag.Bool("verbose", false, "Enable verbose logging to troubleshoot and diagnose issues.")
	pflag.Bool("proxy_metrics", false, "Proxy metrics requests to the upstream Unleash server (ensure that the correct token is provided in the authorization header).")

	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}
