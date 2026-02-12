package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Iandenh/overleash/overleash"
	"github.com/Iandenh/overleash/unleashengine"
	"github.com/charmbracelet/log"
)

func (c *Server) registerFrontendApi(s *http.ServeMux) {
	s.Handle("GET /api/frontend", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Overleash.LockMutex.RLock()
		defer c.Overleash.LockMutex.RUnlock()

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		ctx := createContextFromGetRequest(r)

		result, err := c.featureEnvironmentFromRequest(r).Engine().ResolveAll(ctx, false)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		resultJson, _ := json.Marshal(result)

		w.Write(resultJson)
	}))

	s.Handle("POST /api/frontend", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Overleash.LockMutex.RLock()
		defer c.Overleash.LockMutex.RUnlock()

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		ctx, err := createContextFromPostRequest(r)

		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		result, err := c.featureEnvironmentFromRequest(r).Engine().ResolveAll(ctx, false)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		resultJson, _ := json.Marshal(result)
		w.Write(resultJson)
	}))

	s.Handle("GET /api/frontend/all", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Overleash.LockMutex.RLock()
		defer c.Overleash.LockMutex.RUnlock()

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		ctx := createContextFromGetRequest(r)

		result, err := c.featureEnvironmentFromRequest(r).Engine().ResolveAll(ctx, true)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		resultJson, _ := json.Marshal(result)

		w.Write(resultJson)
	}))

	s.Handle("POST /api/frontend/features/{featureName}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Overleash.LockMutex.RLock()
		defer c.Overleash.LockMutex.RUnlock()

		featureName := r.PathValue("featureName")

		ctx, err := createContextFromPostRequest(r)

		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		result, err := c.featureEnvironmentFromRequest(r).Engine().Resolve(ctx, featureName)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resultJson, _ := json.Marshal(result)

		w.Write(resultJson)
	}))

	s.Handle("GET /api/frontend/features/{featureName}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Overleash.LockMutex.RLock()
		defer c.Overleash.LockMutex.RUnlock()

		featureName := r.PathValue("featureName")

		ctx := createContextFromGetRequest(r)

		result, err := c.featureEnvironmentFromRequest(r).Engine().Resolve(ctx, featureName)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resultJson, _ := json.Marshal(result)

		w.Write(resultJson)
	}))

	s.Handle("POST /api/frontend/client/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		var metric overleash.MetricsData
		err := decoder.Decode(&metric)

		if err != nil {
			http.Error(w, "Error parsing json", http.StatusBadRequest)
			return
		}

		metric.Environment = c.featureEnvironmentFromRequest(r).Environment()

		c.Overleash.AddMetric(&metric)

		w.WriteHeader(http.StatusOK)
	}))

	s.Handle("POST /api/frontend/client/register", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func createContextFromGetRequest(r *http.Request) *unleashengine.Context {
	properties := make(map[string]any)

	ctx := &unleashengine.Context{}

	for k := range r.URL.Query() {
		switch k {
		case "userId":
			ctx.UserId = getQuery(r, "userId")
		case "environment":
			ctx.Environment = getQuery(r, "environment")
		case "appName":
			ctx.AppName = getQuery(r, "appName")
		case "sessionId":
			ctx.SessionId = getQuery(r, "sessionId")
		case "currentTime":
			ctx.CurrentTime = getQuery(r, "currentTime")
		case "remoteAddress":
			ctx.RemoteAddress = getQuery(r, "remoteAddress")
		default:
			if strings.Contains(k, "properties[") {
				key := strings.Split(k, "properties[")[1]
				key = strings.Split(key, "]")[0]
				properties[key] = *getQuery(r, k)
			} else {
				properties[k] = r.URL.Query().Get(k)
			}
		}
	}

	clean := make(map[string]string)
	for k, v := range properties {
		if v != nil {
			clean[k] = v.(string)
		}
	}

	ctx.Properties = clean

	return ctx
}

func createContextFromPostRequest(r *http.Request) (*unleashengine.Context, error) {
	m := map[string]any{}

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&m)

	if err != nil {
		return nil, err
	}
	properties := make(map[string]any)

	ctx := &unleashengine.Context{}

	for k, v := range m {
		switch k {
		case "userId":
			ctx.UserId = getData(v)
		case "environment":
			ctx.Environment = getData(v)
		case "appName":
			ctx.AppName = getData(v)
		case "sessionId":
			ctx.SessionId = getData(v)
		case "currentTime":
			ctx.CurrentTime = getData(v)
		case "remoteAddress":
			ctx.RemoteAddress = getData(v)
		case "properties":
			properties = v.(map[string]any)
		default:
			properties[k] = v
		}
	}

	clean := make(map[string]string)
	for k, v := range properties {
		switch val := v.(type) {
		case string:
			clean[k] = val
		}
	}

	ctx.Properties = clean

	return ctx, nil
}

func getQuery(r *http.Request, name string) *string {
	if !r.URL.Query().Has(name) {
		return nil
	}

	result := r.URL.Query().Get(name)

	return &result
}

func getData(data any) *string {
	var result string

	switch v := data.(type) {
	case string:
		result = v
	case float64:
		result = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		result = strconv.FormatBool(v)
	case nil:
		return nil
	default:
		return nil
	}

	return &result
}
