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

func run(ctx context.Context) {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	tokens := strings.Split(viper.GetString("token"), ",")
	reload := viper.GetInt("reload")
	listenAddress := viper.GetString("listen_address")
	proxyMetrics := viper.GetBool("proxy_metrics")
	registerTokens := viper.GetBool("register")
	headless := viper.GetBool("headless")

	upstream := viper.GetString("upstream")

	if upstream == "" {
		upstream = viper.GetString("url")
		if upstream != "" {
			log.Warn("The 'url' flag is deprecated. Please use the 'upstream' flag instead.")
		}
	}

	o := overleash.NewOverleash(upstream, tokens, reload)
	o.Start(ctx, registerTokens)

	server.New(o, listenAddress, proxyMetrics, ctx, headless).Start()
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
	pflag.String("listen_address", ":5433", "Address to listen on for incoming connections. Can be just a port (e.g. ':5433'), an IP with port (e.g. '127.0.0.1:5433'), or '0.0.0.0:5433' to listen on all interfaces.")
	pflag.Int("reload", 0, "Reload frequency in minutes for refreshing feature flag configuration (0 disables automatic reloading).")
	pflag.Bool("verbose", false, "Enable verbose logging to troubleshoot and diagnose issues.")
	pflag.Bool("proxy_metrics", false, "Proxy metrics requests to the upstream Unleash server (ensure that the correct token is provided in the authorization header).")
	pflag.Bool("register", false, "Whether to register itself to the connected Unleash server.")
	pflag.Bool("headless", false, "Whether to not register the dashboard api.")

	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}
