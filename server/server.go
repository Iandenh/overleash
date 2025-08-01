package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/CAFxX/httpcompression"
	"github.com/Iandenh/overleash/overleash"
	"github.com/a-h/templ"
	"github.com/charmbracelet/log"
	"github.com/rs/cors"
	"io/fs"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var (
	//go:embed static
	staticFiles embed.FS
)

const maxBodySize = 1 * 1024 * 1024 // 1 MiB

type Config struct {
	Overleash    *overleash.OverleashContext
	port         int
	proxyMetrics bool
	ctx          context.Context
}

func New(config *overleash.OverleashContext, port int, proxyMetrics bool, ctx context.Context) *Config {
	return &Config{
		Overleash:    config,
		port:         port,
		proxyMetrics: proxyMetrics,
		ctx:          ctx,
	}
}

func (c *Config) Start() {
	s := http.NewServeMux()

	var staticFS = fs.FS(staticFiles)
	htmlContent, err := fs.Sub(staticFS, "static")

	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(htmlContent))

	s.Handle("/static/", cacheControlMiddleware(http.StripPrefix("/static/", fileServer)))

	c.registerClientApi(s)
	c.registerFrontendApi(s)

	s.HandleFunc("/", func(w http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodDelete {
			c.Overleash.DeleteAllOverride()
		}

		renderFeatures(w, request, c.Overleash)
	})

	s.HandleFunc("POST /override/constrain/{key}/{enabled}", func(w http.ResponseWriter, request *http.Request) {
		key := request.PathValue("key")
		enabled := request.PathValue("enabled")
		flag, err := c.Overleash.ActiveFeatureEnvironment().FeatureFile().Features.Get(key)

		if err != nil {
			http.Error(w, "Feature not found", http.StatusNotFound)
			return
		}

		request.Body = http.MaxBytesReader(w, request.Body, maxBodySize)

		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()

		var constrain overleash.Constraint
		err = decoder.Decode(&constrain)
		if err != nil {
			http.Error(w, "Error parsing json", http.StatusBadRequest)
			return
		}

		c.Overleash.AddOverrideConstraint(key, enabled == "true", constrain)

		templ.Handler(feature(flag, c.Overleash, false)).ServeHTTP(w, request)
	})

	s.HandleFunc("POST /override/{key}/{enabled}", func(w http.ResponseWriter, request *http.Request) {
		key := request.PathValue("key")
		enabled := request.PathValue("enabled")
		flag, err := c.Overleash.ActiveFeatureEnvironment().FeatureFile().Features.Get(key)

		if err != nil {
			http.Error(w, "Feature not found", http.StatusNotFound)
			return
		}

		c.Overleash.AddOverride(key, enabled == "true")

		templ.Handler(feature(flag, c.Overleash, false)).ServeHTTP(w, request)
	})

	s.HandleFunc("DELETE /override/{key}", func(w http.ResponseWriter, request *http.Request) {
		key := request.PathValue("key")

		flag, err := c.Overleash.ActiveFeatureEnvironment().FeatureFile().Features.Get(key)

		if err != nil {
			http.Error(w, "Feature not found", http.StatusNotFound)
			return
		}

		c.Overleash.DeleteOverride(key)

		templ.Handler(feature(flag, c.Overleash, false)).ServeHTTP(w, request)
	})

	s.HandleFunc("POST /dashboard/refresh", func(w http.ResponseWriter, request *http.Request) {
		err := c.Overleash.RefreshFeatureFiles()

		if err != nil {
			http.Error(w, "Failed to refresh feature files", http.StatusInternalServerError)
			return
		}

		updateRequestUrlFromHeader(w, request)

		renderFeatures(w, request, c.Overleash)
	})

	s.HandleFunc("POST /dashboard/search", func(w http.ResponseWriter, request *http.Request) {
		list := search(request, c.Overleash)

		w.Header().Set("HX-Replace-Url", list.url)

		templ.Handler(featureTemplate(list, c.Overleash)).ServeHTTP(w, request)
	})

	s.HandleFunc("POST /dashboard/changeRemote", func(w http.ResponseWriter, request *http.Request) {
		err := request.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusUnprocessableEntity)
			return
		}

		idx, err := strconv.Atoi(request.Form.Get("remote"))

		if err != nil {
			http.Error(w, "Invalid remote index", http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
		}

		err = c.Overleash.SetFeatureFileIdx(idx)

		if err != nil {
			http.Error(w, "Failed to load remote", http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		updateRequestUrlFromHeader(w, request)

		renderFeatures(w, request, c.Overleash)
	})

	s.HandleFunc("POST /dashboard/pause", func(w http.ResponseWriter, request *http.Request) {
		c.Overleash.SetPaused(true)

		updateRequestUrlFromHeader(w, request)

		renderFeatures(w, request, c.Overleash)
	})

	s.HandleFunc("POST /dashboard/unpause", func(w http.ResponseWriter, request *http.Request) {
		c.Overleash.SetPaused(false)

		updateRequestUrlFromHeader(w, request)

		renderFeatures(w, request, c.Overleash)
	})

	s.HandleFunc("GET /dashboard/feature/{key}", func(w http.ResponseWriter, request *http.Request) {
		key := request.PathValue("key")

		flag, err := c.Overleash.ActiveFeatureEnvironment().FeatureFile().Features.Get(key)

		if err != nil {
			http.Error(w, "Feature not found", http.StatusNotFound)
			return
		}

		showDetails := request.URL.Query().Get("details") != ""

		templ.Handler(feature(flag, c.Overleash, showDetails)).ServeHTTP(w, request)
	})

	s.HandleFunc("GET /dashboard/lastSync", func(w http.ResponseWriter, request *http.Request) {
		templ.Handler(lastSync(c.Overleash.LastSync())).ServeHTTP(w, request)
	})

	s.HandleFunc("GET /health", func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := map[string]string{"status": "ok"}
		json.NewEncoder(w).Encode(status)
	})

	handler := cors.AllowAll().Handler(s)
	compress, _ := httpcompression.DefaultAdapter()

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", c.port),
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		Handler:           compress(handler),
	}

	go func() {
		log.Debugf("Starting server on port: %d", c.port)
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
