package server

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/CAFxX/httpcompression"
	"github.com/Iandenh/overleash/overleash"
	"github.com/a-h/templ"
	"github.com/charmbracelet/log"
	"github.com/rs/cors"
)

var (
	//go:embed static
	staticFiles embed.FS
)

const maxBodySize = 1 * 1024 * 1024 // 1 MiB

type Config struct {
	Overleash     *overleash.OverleashContext
	listenAddress string
	ctx           context.Context
	headless      bool
}

func New(config *overleash.OverleashContext, listenAddress string, ctx context.Context, headless bool) *Config {
	return &Config{
		Overleash:     config,
		listenAddress: listenAddress,
		ctx:           ctx,
		headless:      headless,
	}
}

func (c *Config) Start() {
	s := http.NewServeMux()

	if !c.headless {
		var staticFS = fs.FS(staticFiles)
		htmlContent, err := fs.Sub(staticFS, "static")

		if err != nil {
			panic(err)
		}

		fileServer := http.FileServer(http.FS(htmlContent))
		s.Handle("/static/", cacheControlMiddleware(http.StripPrefix("/static/", fileServer)))
	}

	c.registerClientApi(s)
	c.registerFrontendApi(s)

	if !c.headless {
		c.registerDashboardApi(s)
	}

	s.HandleFunc("GET /health", func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := map[string]string{"status": "ok"}
		json.NewEncoder(w).Encode(status)
	})

	handler := cors.AllowAll().Handler(s)
	compress, _ := httpcompression.DefaultAdapter()

	httpServer := &http.Server{
		Addr:              c.listenAddress,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		Handler:           compress(handler),
	}

	go func() {
		log.Debugf("Starting server on port: %s", c.listenAddress)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Error(err)
			panic(err)
		}
	}()

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
	}()
	wg.Wait()
}

func updateRequestUrlFromHeader(w http.ResponseWriter, request *http.Request) {
	if u := request.Header.Get("hx-current-url"); u != "" {
		parsedUrl, err := url.Parse(u)
		if err != nil {
			http.Error(w, "Failed to parse url", http.StatusBadRequest)
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
