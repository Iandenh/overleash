package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/Iandenh/overleash/overleash"
	unleash "github.com/Unleash/unleash-client-go/v4/api"
	"github.com/a-h/templ"
	"github.com/charmbracelet/log"
	"github.com/rs/cors"
	"io/fs"
	"net/http"
	"strconv"
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
}

func New(config *overleash.OverleashContext, port int, proxyMetrics bool) *Config {
	return &Config{
		Overleash:    config,
		port:         port,
		proxyMetrics: proxyMetrics,
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

	s.Handle("/", NewFeaturesHandler(c.Overleash))
	s.Handle("/static/", http.StripPrefix("/static/", fileServer))

	middleware := createNewDynamicModeMiddleware(c.Overleash)

	c.registerClientApi(s, middleware)
	c.registerFrontendApi(s, middleware)

	s.HandleFunc("POST /override/constrain/{key}/{enabled}", func(w http.ResponseWriter, request *http.Request) {
		key := request.PathValue("key")
		enabled := request.PathValue("enabled")
		flag, err := c.Overleash.FeatureFile().Features.Get(key)

		if err != nil && !c.Overleash.IsDynamicMode() {
			http.Error(w, "Feature not found", http.StatusNotFound)
			return
		}

		request.Body = http.MaxBytesReader(w, request.Body, maxBodySize)

		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()

		var constrain unleash.Constraint
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
		flag, err := c.Overleash.FeatureFile().Features.Get(key)

		if err != nil && !c.Overleash.IsDynamicMode() {
			http.Error(w, "Feature not found", http.StatusNotFound)
			return
		}

		c.Overleash.AddOverride(key, enabled == "true")

		templ.Handler(feature(flag, c.Overleash, false)).ServeHTTP(w, request)
	})

	s.HandleFunc("DELETE /override/{key}", func(w http.ResponseWriter, request *http.Request) {
		key := request.PathValue("key")

		flag, err := c.Overleash.FeatureFile().Features.Get(key)

		if err != nil && !c.Overleash.IsDynamicMode() {
			http.Error(w, "Feature not found", http.StatusNotFound)
			return
		}

		c.Overleash.DeleteOverride(key)

		templ.Handler(feature(flag, c.Overleash, false)).ServeHTTP(w, request)
	})

	s.HandleFunc("POST /refresh", func(w http.ResponseWriter, request *http.Request) {
		err := c.Overleash.RefreshFeatureFiles()

		if err != nil {
			http.Error(w, "Failed to refresh feature files", http.StatusInternalServerError)
			return
		}

		renderFeatures(w, request, c.Overleash)
	})

	s.HandleFunc("POST /search", func(w http.ResponseWriter, request *http.Request) {
		list := search(request, c.Overleash)

		w.Header().Set("HX-Replace-Url", list.url)

		templ.Handler(featureTemplate(list, c.Overleash)).ServeHTTP(w, request)
	})

	s.HandleFunc("POST /pause", func(w http.ResponseWriter, request *http.Request) {
		c.Overleash.SetPaused(true)

		renderFeatures(w, request, c.Overleash)
	})

	s.HandleFunc("POST /changeRemote", func(w http.ResponseWriter, request *http.Request) {
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

		renderFeatures(w, request, c.Overleash)
	})

	s.HandleFunc("POST /unpause", func(w http.ResponseWriter, request *http.Request) {
		c.Overleash.SetPaused(false)

		renderFeatures(w, request, c.Overleash)
	})

	s.HandleFunc("GET /feature/{key}", func(w http.ResponseWriter, request *http.Request) {
		key := request.PathValue("key")

		flag, err := c.Overleash.FeatureFile().Features.Get(key)

		if err != nil {
			http.Error(w, "Feature not found", http.StatusNotFound)
			return
		}

		showDetails := request.URL.Query().Get("details") != ""

		templ.Handler(feature(flag, c.Overleash, showDetails)).ServeHTTP(w, request)
	})

	s.HandleFunc("GET /lastSync", func(w http.ResponseWriter, request *http.Request) {
		templ.Handler(lastSync(c.Overleash.LastSync())).ServeHTTP(w, request)
	})

	s.HandleFunc("GET /health", func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ok"))
	})

	handler := cors.AllowAll().Handler(s)
	log.Debugf("Starting server on port: %d", c.port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", c.port), handler); err != nil {
		log.Error(err)
		panic(err)
	}
}

type FeaturesHandler struct {
	GetContext func() *overleash.OverleashContext
}

func NewFeaturesHandler(o *overleash.OverleashContext) FeaturesHandler {
	ContextGetter := func() *overleash.OverleashContext {
		return o
	}
	return FeaturesHandler{
		GetContext: ContextGetter,
	}
}

func (fh FeaturesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	o := fh.GetContext()

	if r.Method == http.MethodDelete {
		o.DeleteAllOverride()
	}

	renderFeatures(w, r, o)
}

func renderFeatures(w http.ResponseWriter, r *http.Request, o *overleash.OverleashContext) {
	list := search(r, o)

	w.Header().Set("HX-Replace-Url", list.url)

	templ.Handler(features(list, o)).ServeHTTP(w, r)
}
