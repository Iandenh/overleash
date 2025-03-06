package server

import (
	"encoding/json"
	"github.com/Iandenh/overleash/proxy"
	"github.com/Iandenh/overleash/unleashengine"
	"github.com/charmbracelet/log"
	"net/http"
	"strconv"
	"strings"
)

func (c *Config) registerFrontendApi(s *http.ServeMux) {
	s.Handle("GET /api/frontend", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Overleash.LockMutex.RLock()
		defer c.Overleash.LockMutex.RUnlock()

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		ctx := createContextFromGetRequest(r)

		resolvedToggles, ok := resolveAll(c.Overleash.Engine(), ctx)
		if !ok {
			return
		}

		result := frontendFromYggdrasil(resolvedToggles, false)
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

		resolvedToggles, ok := resolveAll(c.Overleash.Engine(), ctx)
		if !ok {
			return
		}

		result := frontendFromYggdrasil(resolvedToggles, false)
		resultJson, _ := json.Marshal(result)

		w.Write(resultJson)
	}))

	s.Handle("GET /api/frontend/all", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Overleash.LockMutex.RLock()
		defer c.Overleash.LockMutex.RUnlock()

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		ctx := createContextFromGetRequest(r)

		resolvedToggles, ok := resolveAll(c.Overleash.Engine(), ctx)
		if !ok {
			return
		}

		result := frontendFromYggdrasil(resolvedToggles, true)
		resultJson, _ := json.Marshal(result)

		w.Write(resultJson)
	}))

	s.Handle("GET /api/frontend/features/{featureName}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Overleash.LockMutex.RLock()
		defer c.Overleash.LockMutex.RUnlock()

		featureName := r.PathValue("featureName")

		ctx := createContextFromGetRequest(r)

		resolved, ok := resolve(c.Overleash.Engine(), ctx, featureName)

		if !ok {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		evaluated := EvaluatedToggle{
			Name:    featureName,
			Enabled: resolved.Enabled,
			Variant: EvaluatedVariant{
				Name:              resolved.Variant.Name,
				Enabled:           resolved.Variant.Enabled,
				Payload:           resolved.Variant.Payload,
				FeatureEnabled:    resolved.Variant.FeatureEnabled,
				OldFeatureEnabled: resolved.Variant.FeatureEnabled,
			},
			ImpressionData: resolved.ImpressionData,
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resultJson, _ := json.Marshal(evaluated)

		w.Write(resultJson)
	}))

	s.Handle("POST /api/frontend/client/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.proxyMetrics == false {
			w.WriteHeader(http.StatusOK)

			return
		}
		p := proxy.New(c.Overleash.Upstream())

		err := p.ProxyRequest(w, r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))

	s.Handle("POST /api/frontend/client/register", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func resolveAll(engine *unleashengine.UnleashEngine, ctx *unleashengine.Context) (map[string]ResolvedToggle, bool) {
	var apiResponse apiResponse[map[string]ResolvedToggle]
	err := json.Unmarshal(engine.ResolveAll(ctx), &apiResponse)
	if err != nil {
		log.Errorf("Error unmarshalling JSON: %s", err)
		return map[string]ResolvedToggle{}, false
	}
	return apiResponse.Value, true
}

func resolve(engine *unleashengine.UnleashEngine, ctx *unleashengine.Context, featureName string) (ResolvedToggle, bool) {
	var apiResponse apiResponse[ResolvedToggle]
	err := json.Unmarshal(engine.Resolve(ctx, featureName), &apiResponse)
	if err != nil {
		log.Errorf("Error unmarshalling JSON: %s", err)
		return ResolvedToggle{}, false
	}

	if apiResponse.StatusCode == "NotFound" {
		return ResolvedToggle{}, false
	}

	return apiResponse.Value, true
}

type FrontendResult struct {
	Toggles []EvaluatedToggle `json:"toggles"`
}

type EvaluatedToggle struct {
	Name           string           `json:"name"`
	Enabled        bool             `json:"enabled"`
	Variant        EvaluatedVariant `json:"variant"`
	ImpressionData bool             `json:"impressionData"`
}

type EvaluatedVariant struct {
	Name              string         `json:"name"`
	Enabled           bool           `json:"enabled"`
	Payload           VariantPayload `json:"payload"`
	FeatureEnabled    bool           `json:"feature_enabled"`
	OldFeatureEnabled bool           `json:"featureEnabled"`
}

type VariantPayload struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ResolvedToggle struct {
	Enabled        bool    `json:"enabled"`
	Project        string  `json:"project"`
	Variant        Variant `json:"variant"`
	ImpressionData bool    `json:"impressionData"`
}

type Variant struct {
	Name           string         `json:"name"`
	Enabled        bool           `json:"enabled"`
	FeatureEnabled bool           `json:"feature_enabled"`
	Payload        VariantPayload `json:"payload"`
}

type apiResponse[T any] struct {
	StatusCode   string `json:"status_code"`
	Value        T      `json:"value"`
	ErrorMessage string `json:"error_message"`
}

func frontendFromYggdrasil(res map[string]ResolvedToggle, includeAll bool) FrontendResult {
	toggles := make([]EvaluatedToggle, 0)

	for name, resolved := range res {
		if includeAll || resolved.Enabled {
			toggle := EvaluatedToggle{
				Name:    name,
				Enabled: resolved.Enabled,
				Variant: EvaluatedVariant{
					Name:              resolved.Variant.Name,
					Enabled:           resolved.Variant.Enabled,
					Payload:           resolved.Variant.Payload,
					FeatureEnabled:    resolved.Variant.FeatureEnabled,
					OldFeatureEnabled: resolved.Variant.FeatureEnabled,
				},
				ImpressionData: resolved.ImpressionData,
			}
			toggles = append(toggles, toggle)
		}
	}

	return FrontendResult{Toggles: toggles}
}

func createContextFromGetRequest(r *http.Request) *unleashengine.Context {
	properties := make(map[string]interface{})

	ctx := &unleashengine.Context{}

	for k, _ := range r.URL.Query() {
		switch k {
		case "userId":
			ctx.UserID = getQuery(r, "userId")
		case "environment":
			ctx.Environment = getQuery(r, "environment")
		case "appName":
			ctx.AppName = getQuery(r, "appName")
		case "sessionId":
			ctx.SessionID = getQuery(r, "sessionId")
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

	ctx.Properties = &clean

	return ctx
}

func createContextFromPostRequest(r *http.Request) (*unleashengine.Context, error) {
	m := map[string]interface{}{}

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&m)

	if err != nil {
		return nil, err
	}
	properties := make(map[string]interface{})

	ctx := &unleashengine.Context{}

	for k, v := range m {
		switch k {
		case "userId":
			ctx.UserID = getData(v)
		case "environment":
			ctx.Environment = getData(v)
		case "appName":
			ctx.AppName = getData(v)
		case "sessionId":
			ctx.SessionID = getData(v)
		case "currentTime":
			ctx.CurrentTime = getData(v)
		case "remoteAddress":
			ctx.RemoteAddress = getData(v)
		case "properties":
			properties = v.(map[string]interface{})
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

	ctx.Properties = &clean

	return ctx, nil
}

func getQuery(r *http.Request, name string) *string {
	if !r.URL.Query().Has(name) {
		return nil
	}

	result := r.URL.Query().Get(name)

	return &result
}

func getData(data interface{}) *string {
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
