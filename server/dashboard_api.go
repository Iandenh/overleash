package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Iandenh/overleash/overleash"
	"github.com/a-h/templ"
)

func (c *Config) registerDashboardApi(s *http.ServeMux) {
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
}
