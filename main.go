package main

import (
	"github.com/charmbracelet/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"overleash/overleash"
	"overleash/server"
	"strings"
)

func main() {
	log.Info("Starting Overleash")
	initConfig()

	if viper.GetBool("verbose") {
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug logging enabled")
		log.Debug(viper.AllSettings())
	}

	tokens := strings.Split(viper.GetString("token"), ",")
	reload := viper.GetInt("reload")
	port := viper.GetInt("port")
	dynamicMode := viper.GetBool("dynamic_mode")

	o := overleash.NewOverleash(viper.GetString("url"), tokens, dynamicMode)
	o.Start(reload)

	server.New(o, port).Start()
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
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}
