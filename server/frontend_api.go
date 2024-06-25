package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"overleash/unleashengine"
	"strings"
)

func (c *Config) registerFrontendApi(s *http.ServeMux) {
	s.HandleFunc("GET /api/frontend", func(w http.ResponseWriter, r *http.Request) {
		c.Overleash.LockMutex.RLock()
		defer c.Overleash.LockMutex.RUnlock()

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)

		ctx := createContextFromRequest(r)

		apiResponse, done := resolveAll(c.Overleash.Engine(), ctx)
		if done {
			return
		}

		result := frontendFromYggdrasil(apiResponse.Value, false)
		resultJson, _ := json.Marshal(result)

		w.Write(resultJson)
	})

	s.HandleFunc("GET /api/frontend/all", func(w http.ResponseWriter, r *http.Request) {
		c.Overleash.LockMutex.RLock()
		defer c.Overleash.LockMutex.RUnlock()

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)

		ctx := createContextFromRequest(r)

		apiResponse, done := resolveAll(c.Overleash.Engine(), ctx)
		if done {
			return
		}

		result := frontendFromYggdrasil(apiResponse.Value, true)
		resultJson, _ := json.Marshal(result)

		w.Write(resultJson)
	})

	s.HandleFunc("POST /api/frontend/client/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	s.HandleFunc("POST /api/frontend/client/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
}

func resolveAll(engine *unleashengine.UnleashEngine, ctx *unleashengine.Context) (ApiResponse, bool) {
	var apiResponse ApiResponse
	err := json.Unmarshal(engine.ResolveAll(ctx), &apiResponse)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return ApiResponse{}, true
	}
	return apiResponse, false
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
