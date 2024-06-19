package server

import (
	"embed"
	"fmt"
	"github.com/a-h/templ"
	"github.com/rs/cors"
	"io/fs"
	"net/http"
	"overleash/overleash"
	"strconv"
)

var (
	//go:embed static
	staticFiles embed.FS
)

type Config struct {
	Overleash *overleash.OverleashContext
	port      int
}

func New(config *overleash.OverleashContext, port int) *Config {
	return &Config{
		Overleash: config,
		port:      port,
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

	c.registerClientApi(s)
	c.registerFrontendApi(s)

	s.HandleFunc("POST /override/{key}/{enabled}", func(w http.ResponseWriter, request *http.Request) {
		key := request.PathValue("key")
		enabled := request.PathValue("enabled")
		flag, err := c.Overleash.FeatureFile().Features.Get(key)

		if err != nil {
			w.WriteHeader(404)
			return
		}

		c.Overleash.AddOverride(key, enabled == "true")

		templ.Handler(feature(flag, c.Overleash)).ServeHTTP(w, request)
	})

	s.HandleFunc("DELETE /override/{key}", func(w http.ResponseWriter, request *http.Request) {
		key := request.PathValue("key")

		flag, err := c.Overleash.FeatureFile().Features.Get(key)

		if err != nil {
			w.WriteHeader(404)
			return
		}

		c.Overleash.DeleteOverride(key)

		templ.Handler(feature(flag, c.Overleash)).ServeHTTP(w, request)
	})

	s.HandleFunc("POST /refresh", func(w http.ResponseWriter, request *http.Request) {
		err := c.Overleash.RefreshDatafiles()

		if err != nil {
			w.WriteHeader(404)
			return
		}

		templ.Handler(features(c.Overleash)).ServeHTTP(w, request)
	})

	s.HandleFunc("POST /search", func(w http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		search := request.Form.Get("search")

		flags := fuzzyFeatureFlags(search, c.Overleash)

		templ.Handler(featureTemplate(flags, c.Overleash)).ServeHTTP(w, request)
	})

	s.HandleFunc("POST /pause", func(w http.ResponseWriter, request *http.Request) {
		c.Overleash.SetPaused(true)

		templ.Handler(features(c.Overleash)).ServeHTTP(w, request)
	})

	s.HandleFunc("POST /changeRemote", func(w http.ResponseWriter, request *http.Request) {
		request.ParseForm()

		idx, err := strconv.Atoi(request.Form.Get("remote"))

		if err != nil {
			w.WriteHeader(400)
			return
		}

		err = c.Overleash.SetDataFileIdx(idx)

		if err != nil {
			w.WriteHeader(400)
			return
		}

		templ.Handler(features(c.Overleash)).ServeHTTP(w, request)
	})

	s.HandleFunc("POST /unpause", func(w http.ResponseWriter, request *http.Request) {
		c.Overleash.SetPaused(false)

		templ.Handler(features(c.Overleash)).ServeHTTP(w, request)
	})

	s.HandleFunc("GET /lastSync", func(w http.ResponseWriter, request *http.Request) {
		templ.Handler(lastSync(c.Overleash.LastSync())).ServeHTTP(w, request)
	})

	handler := cors.AllowAll().Handler(s)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", c.port), handler); err != nil {
		panic(err)
	}
}

type FeaturesHandler struct {
	GetContext func() *overleash.OverleashContext
}

func NewFeaturesHandler(o *overleash.OverleashContext) FeaturesHandler {
	// Replace this in-memory function with a call to a database.
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

	templ.Handler(features(o)).ServeHTTP(w, r)
}
