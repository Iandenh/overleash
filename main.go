package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

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
	listenAddress := viper.GetString("listen_address")
	registerMetrics := viper.GetBool("register_metrics")
	registerTokens := viper.GetBool("register")
	useDeltaApi := viper.GetBool("delta")
	streamer := viper.GetBool("streamer")
	headless := viper.GetBool("headless")
	frontendApiEnabled := viper.GetBool("enable_frontend_api")
	envFromToken := viper.GetBool("env_from_token")
	prometheusMetrics := viper.GetBool("prometheus_metrics")
	prometheusMetricsPort := viper.GetInt("prometheus_metrics_port")

	upstream := viper.GetString("upstream")

	if upstream == "" {
		upstream = viper.GetString("url")
		if upstream != "" {
			log.Warn("The 'url' flag is deprecated. Please use the 'upstream' flag instead.")
		}
	}

	o := overleash.NewOverleash(upstream, tokens, parseReload(), streamer, frontendApiEnabled)
	o.Start(ctx, registerMetrics, registerTokens, useDeltaApi)

	server.New(o, listenAddress, ctx, headless, envFromToken, prometheusMetrics, prometheusMetricsPort).Start()
}

func parseReload() time.Duration {
	reloadStr := viper.GetString("reload")

	r, err := time.ParseDuration(reloadStr)
	if err != nil {
		// try interpreting as minutes if it's just a number
		if num, convErr := strconv.Atoi(reloadStr); convErr == nil {
			r = time.Duration(num) * time.Minute
		} else {
			panic(err) // real parse error
		}
	}

	return r
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
	pflag.Bool("register_metrics", false, "Register metrics")
	pflag.Bool("register", false, "Whether to register itself to the connected Unleash server.")
	pflag.Bool("headless", false, "Whether to not register the dashboard api.")
	pflag.Bool("streamer", false, "Whether this instance streams the delta events.")
	pflag.Bool("enable_frontend_api", true, "Whether to enable the frontend API.")
	pflag.Bool("delta", false, "Whether to enable the frontend API.")
	pflag.Bool("env_from_token", true, "Whether to resolve the environment from the client token in the Authorization header instead of using the configured environment.")
	pflag.Bool("prometheus_metrics", false, "Whether to collect prometheus metrics from the server.")
	pflag.Int("prometheus_metrics_port", 9100, "Which port to expose Prometheus metrics.")

	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}
