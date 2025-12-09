package config

import (
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	// Core
	URL      string `mapstructure:"url"`
	Upstream string `mapstructure:"upstream"`
	Token    string `mapstructure:"token"`

	// Network / Routing
	BasePath string `mapstructure:"base_path"`

	// Server
	ListenAddress string `mapstructure:"listen_address"`
	Reload        string `mapstructure:"reload"`

	Backup bool `mapstructure:"backup"`

	// Logging & Metrics
	Verbose           bool `mapstructure:"verbose"`
	RegisterMetrics   bool `mapstructure:"register_metrics"`
	PrometheusMetrics bool `mapstructure:"prometheus_metrics"`
	PrometheusPort    int  `mapstructure:"prometheus_metrics_port"`

	// Unleash-specific
	Register       bool `mapstructure:"register"`
	Headless       bool `mapstructure:"headless"`
	Streamer       bool `mapstructure:"streamer"`
	EnableFrontend bool `mapstructure:"enable_frontend_api"`
	Delta          bool `mapstructure:"delta"`
	EnvFromToken   bool `mapstructure:"env_from_token"`
	Webhook        bool `mapstructure:"webhook"`

	// Storage
	Storage string `mapstructure:"storage"`

	// Redis
	RedisAddr      string `mapstructure:"redis_address"`
	RedisPassword  string `mapstructure:"redis_password"`
	RedisDB        int    `mapstructure:"redis_db"`
	RedisChannel   string `mapstructure:"redis_channel"`
	RedisSentinel  bool   `mapstructure:"redis_sentinel"`
	RedisMaster    string `mapstructure:"redis_master"`
	RedisSentinels string `mapstructure:"redis_sentinels"`
}

func InitConfig() (*Config, error) {
	viper.SetEnvPrefix("OVERLEASH")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	pflag.String("url", "", "DEPRECATED: Unleash URL (e.g. https://unleash.my-site.com) without /api. Use --upstream instead.")
	pflag.String("upstream", "", "Unleash upstream URL to load feature flags (e.g. https://unleash.my-site.com) without /api, can be an Unleash instance or Unleash Edge.")
	pflag.String("token", "", "Comma-separated Unleash client token(s) to fetch feature flag configurations.")
	pflag.String("base_path", "", "Base URL path if running behind an ingress with a prefix (e.g. /overleash).")
	pflag.String("listen_address", ":5433", "Address to listen on for incoming connections. Can be just a port (e.g. ':5433'), an IP with port (e.g. '127.0.0.1:5433'), or '0.0.0.0:5433' to listen on all interfaces.")
	pflag.String("reload", "0", "Reload frequency in minutes for refreshing feature flag configuration (0 disables automatic reloading).")
	pflag.Bool("verbose", false, "Enable verbose logging to troubleshoot and diagnose issues.")
	pflag.Bool("register_metrics", false, "Register metrics")
	pflag.Bool("register", false, "Whether to register itself to the connected Unleash server.")
	pflag.Bool("headless", false, "Whether to not register the dashboard api.")
	pflag.Bool("streamer", false, "Whether this instance streams the delta events.")
	pflag.Bool("enable_frontend_api", true, "Whether to enable the frontend API.")
	pflag.Bool("delta", false, "Whether to to use the upstream delta streaming API.")
	pflag.Bool("env_from_token", false, "Whether to resolve the environment from the client token in the Authorization header instead of using the configured environment.")
	pflag.Bool("prometheus_metrics", false, "Whether to collect prometheus metrics from the server.")
	pflag.Int("prometheus_metrics_port", 9100, "Which port to expose Prometheus metrics.")
	pflag.Bool("webhook", false, "Whether to expose webhook that will refresh the flags.")
	pflag.Bool("backup", true, "Whether backup feature file in storage.")

	pflag.String("storage", "file", "Storage backend: file or redis")

	pflag.String("redis_address", "localhost:6379", "Redis address (host:port)")
	pflag.String("redis_password", "", "Redis password")
	pflag.Int("redis_db", 0, "Redis DB number")
	pflag.String("redis_channel", "overrides-updates", "Redis Pub/Sub channel")

	pflag.Bool("redis_sentinel", false, "Use Redis Sentinel")
	pflag.String("redis_master", "mymaster", "Redis master name (for Sentinel)")
	pflag.String("redis_sentinels", "", "Comma separated list of sentinel addresses (host:port)")

	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) ParseReload() time.Duration {
	r, err := time.ParseDuration(c.Reload)
	if err != nil {
		// try interpreting as minutes if it's just a number
		if num, convErr := strconv.Atoi(c.Reload); convErr == nil {
			r = time.Duration(num) * time.Minute
		} else {
			panic(err) // real parse error
		}
	}

	return r
}

func (c *Config) Tokens() []string {
	return strings.Split(c.Token, ",")
}

// CleanBasePath ensures the path starts with / and does not end with /
// This makes it safe for http.StripPrefix
func (c *Config) CleanBasePath() string {
	if c.BasePath == "" || c.BasePath == "/" {
		return ""
	}
	// Ensure leading slash
	path := c.BasePath
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// Remove trailing slash
	return strings.TrimSuffix(path, "/")
}
