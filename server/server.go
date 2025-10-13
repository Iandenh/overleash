package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/CAFxX/httpcompression"
	"github.com/Iandenh/overleash/overleash"
	"github.com/a-h/templ"
	"github.com/charmbracelet/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

var (
	//go:embed static
	staticFiles embed.FS
)

const maxBodySize = 1 * 1024 * 1024 // 1 MiB

type Server struct {
	Overleash *overleash.OverleashContext
	ctx       context.Context
}

func New(config *overleash.OverleashContext, ctx context.Context) *Server {
	return &Server{
		Overleash: config,
		ctx:       ctx,
	}
}

func (c *Server) Start() {
	s := http.NewServeMux()

	if !c.Overleash.Config.Headless {
		var staticFS = fs.FS(staticFiles)
		htmlContent, err := fs.Sub(staticFS, "static")

		if err != nil {
			panic(err)
		}

		fileServer := http.FileServer(http.FS(htmlContent))
		s.Handle("/static/", cacheControlMiddleware(http.StripPrefix("/static/", fileServer)))
	}

	c.registerClientApi(s)
	c.registerEdgeApi(s)

	if c.Overleash.Config.Webhook {
		c.registerWebhookApi(s)
	}

	if c.Overleash.Config.EnableFrontend {
		c.registerFrontendApi(s)
	}

	if c.Overleash.Config.Streamer {
		c.registerDeltaApi(s)
	}

	if !c.Overleash.Config.Headless {
		c.registerDashboardApi(s)
	}

	s.HandleFunc("GET /health", func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := map[string]string{"status": "ok"}
		json.NewEncoder(w).Encode(status)
	})

	handler := cors.AllowAll().Handler(s)
	compress, _ := httpcompression.DefaultAdapter()

	handler = compress(handler)
	if c.Overleash.Config.PrometheusMetrics {
		handler = instrumentHandler(handler)
	}

	httpServer := &http.Server{
		Addr:    c.Overleash.Config.ListenAddress,
		Handler: handler,
	}

	go func() {
		log.Debugf("Starting server on port: %s", c.Overleash.Config.ListenAddress)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Error(err)
			panic(err)
		}
	}()

	var metricsServer *http.Server
	if c.Overleash.Config.PrometheusMetrics == true {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())

		metricsServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", c.Overleash.Config.PrometheusPort),
			Handler: mux,
		}

		go func() {
			log.Debugf("Starting metrics server on port: %d", c.Overleash.Config.PrometheusPort)
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Errorf("metrics server error: %v", err)
			}
		}()
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-c.ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Debug("Shutting down server")
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Errorf("error shutting down http server: %s\n", err)
		}

		if metricsServer != nil {
			if err := metricsServer.Shutdown(shutdownCtx); err != nil {
				log.Errorf("error shutting down metrics server: %s", err)
			}
		}
	}()
	wg.Wait()
}

func updateRequestUrlFromHeader(w http.ResponseWriter, request *http.Request) {
	if u := request.Header.Get("hx-current-url"); u != "" {
		parsedUrl, err := url.Parse(u)
		if err != nil {
			http.Error(w, "Failed to parse url", http.StatusBadRequest)
			return
		}

		request.URL = parsedUrl
	}
}

func renderFeatures(w http.ResponseWriter, r *http.Request, o *overleash.OverleashContext) {
	list := search(r, o)

	w.Header().Set("HX-Replace-Url", list.url)

	templ.Handler(features(list, o, getColorScheme(r))).ServeHTTP(w, r)
}

func getColorScheme(r *http.Request) string {
	cookie, err := r.Cookie("prefers-color-scheme")

	if err != nil {
		return "auto"
	}

	switch cookie.Value {
	case "light":
		return "light"
	case "dark":
		return "dark"
	}

	return "auto"
}
