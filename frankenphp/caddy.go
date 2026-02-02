package frankenphp

import (
	"context"
	"net/http"
	"strings"

	"github.com/Iandenh/overleash/config"
	"github.com/Iandenh/overleash/overleash"
	"github.com/Iandenh/overleash/server"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

const defaultHubURL = "/.well-known/overleash"

func init() {
	caddy.RegisterModule(Overleash{})
	httpcaddyfile.RegisterHandlerDirective("overleash", parseCaddyfile)
	httpcaddyfile.RegisterDirectiveOrder("overleash", "after", "encode")
}

var hub *Overleash

type Overleash struct {
	Upstream        string `json:"upstream,omitempty"`
	Token           string `json:"token,omitempty"`
	Reload          string `json:"reload,omitempty"`
	Verbose         bool   `json:"verbose,omitempty"`
	Register        bool   `json:"register,omitempty"`
	RegisterMetrics bool   `json:"register_metrics,omitempty"`
	Streamer        bool   `json:"streamer,omitempty"`
	Delta           bool   `json:"delta,omitempty"`

	cancel    context.CancelFunc
	overleash *overleash.OverleashContext
	server    http.Handler
	logger    *zap.Logger
}

func (Overleash) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.overleash",
		New: func() caddy.Module { return new(Overleash) },
	}
}

func (o *Overleash) Provision(ctx caddy.Context) (err error) {
	o.logger = ctx.Logger()

	hub = o

	o.overleash = overleash.NewOverleash(&config.Config{
		Upstream:        o.Upstream,
		Token:           o.Token,
		BasePath:        defaultHubURL,
		Reload:          o.Reload,
		Backup:          false,
		Verbose:         o.Verbose,
		RegisterMetrics: o.RegisterMetrics,
		PrometheusPort:  0,
		Register:        false,
		Headless:        false,
		EnableFrontend:  true,
		Streamer:        o.Streamer,
		Delta:           o.Delta,
		Storage:         "file",
	})
	o.overleash.Start(ctx)

	o.server = server.New(o.overleash, ctx).CreateHandler()

	return nil
}

func (o Overleash) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	if !strings.HasPrefix(r.URL.Path, defaultHubURL) {
		return next.ServeHTTP(w, r)
	}

	o.server.ServeHTTP(w, r)

	return nil
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var o Overleash

	return o, o.UnmarshalCaddyfile(h.Dispenser)
}

func (o *Overleash) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "upstream":
				if !d.NextArg() {
					return d.ArgErr()
				}

				o.Upstream = d.Val()
			case "token":
				if !d.NextArg() {
					return d.ArgErr()
				}

				o.Token = d.Val()
			case "reload":
				if !d.NextArg() {
					return d.ArgErr()
				}

				o.Reload = d.Val()
			case "verbose":
				o.Verbose = true
			case "register":
				o.Register = true
			case "register_metrics":
				o.RegisterMetrics = true
			case "streamer":
				o.Streamer = true
			case "delta":
				o.Delta = true
			}
		}
	}

	return nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Overleash)(nil)
	_ caddyhttp.MiddlewareHandler = (*Overleash)(nil)
	_ caddyfile.Unmarshaler       = (*Overleash)(nil)
)
