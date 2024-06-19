package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"overleash/overleash"
	"overleash/server"
	"strings"
)

func main() {
	initConfig()

	tokens := strings.Split(viper.GetString("token"), ",")

	reload := viper.GetInt("reload")
	port := viper.GetInt("port")

	o := overleash.NewOverleash(viper.GetString("url"), tokens)
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
	pflag.Int("reload", 0, "Reload")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}
