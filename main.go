package main

import (
	"context"
	"github.com/Iandenh/overleash/overleash"
	"github.com/Iandenh/overleash/server"
	"github.com/charmbracelet/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"strings"
)

func run(ctx context.Context) {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	tokens := strings.Split(viper.GetString("token"), ",")
	reload := viper.GetInt("reload")
	port := viper.GetInt("port")
	proxyMetrics := viper.GetBool("proxy_metrics")
	dynamicMode := viper.GetBool("dynamic_mode")

	o := overleash.NewOverleash(viper.GetString("url"), tokens, dynamicMode)
	o.Start(reload, ctx)

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

	pflag.String("url", "", "Url to load")
	pflag.String("token", "", "Token to load")
	pflag.String("port", "5433", "Port")
	pflag.Bool("dynamic_mode", false, "If enable dynamic mode")
	pflag.Int("reload", 0, "Reload")
	pflag.Bool("verbose", false, "Verbose mode")
	pflag.Bool("proxy_metrics", false, "Proxy metrics")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}
