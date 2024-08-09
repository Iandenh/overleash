package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"overleash/unleashengine"
	"strings"
)

func (c *Config) registerFrontendApi(s *http.ServeMux, middleware Middleware) {
	s.Handle("GET /api/frontend", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Overleash.LockMutex.RLock()
		defer c.Overleash.LockMutex.RUnlock()

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)

		ctx := createContextFromRequest(r)

		apiResponse, ok := resolveAll(c.Overleash.Engine(), ctx)
		if !ok {
			return
		}

		result := frontendFromYggdrasil(apiResponse.Value, false)
		resultJson, _ := json.Marshal(result)

		w.Write(resultJson)
	})))

	s.Handle("GET /api/frontend/all", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Overleash.LockMutex.RLock()
		defer c.Overleash.LockMutex.RUnlock()

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)

		ctx := createContextFromRequest(r)

		apiResponse, ok := resolveAll(c.Overleash.Engine(), ctx)
		if !ok {
			return
		}

		result := frontendFromYggdrasil(apiResponse.Value, true)
		resultJson, _ := json.Marshal(result)

		w.Write(resultJson)
	})))

	s.Handle("GET /api/frontend/features/{featureName}", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		featureName := r.PathValue("featureName")
		c.Overleash.LockMutex.RLock()
		defer c.Overleash.LockMutex.RUnlock()

		ctx := createContextFromRequest(r)

		resolved, ok := resolve(c.Overleash.Engine(), ctx, featureName)

		if !ok {
			w.WriteHeader(404)

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
		w.WriteHeader(200)

		resultJson, _ := json.Marshal(evaluated)

		w.Write(resultJson)
	})))

	s.Handle("POST /api/frontend/client/metrics", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})))

	s.Handle("POST /api/frontend/client/register", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})))
}

func resolveAll(engine *unleashengine.UnleashEngine, ctx *unleashengine.Context) (ApiResponse, bool) {
	var apiResponse ApiResponse
	err := json.Unmarshal(engine.ResolveAll(ctx), &apiResponse)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return ApiResponse{}, false
	}
	return apiResponse, true
}

func resolve(engine *unleashengine.UnleashEngine, ctx *unleashengine.Context, featureName string) (ResolvedToggle, bool) {
	var resolvedToggle ResolvedToggleResult
	err := json.Unmarshal(engine.Resolve(ctx, featureName), &resolvedToggle)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return ResolvedToggle{}, false
	}

	if resolvedToggle.StatusCode == "NotFound" {
		return ResolvedToggle{}, false
	}

	return resolvedToggle.Value, true
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

type EdgeToken struct {
	Projects []string `json:"projects"`
}

type ApiResponse struct {
	StatusCode   string                    `json:"status_code"`
	Value        map[string]ResolvedToggle `json:"value"`
	ErrorMessage string                    `json:"error_message"`
}

type ResolvedToggleResult struct {
	StatusCode   string         `json:"status_code"`
	Value        ResolvedToggle `json:"value"`
	ErrorMessage string         `json:"error_message"`
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

func createContextFromRequest(r *http.Request) *unleashengine.Context {
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

func getQuery(r *http.Request, name string) *string {
	if !r.URL.Query().Has(name) {
		return nil
	}

	result := r.URL.Query().Get(name)

	return &result
}
