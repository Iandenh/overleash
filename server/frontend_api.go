package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"overleash/unleashengine"
)

var engine *unleashengine.UnleashEngine

func init() {
	engine = unleashengine.NewUnleashEngine()
}

func (c *Config) registerFrontendApi(s *http.ServeMux) {
	s.HandleFunc("GET /api/frontend", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)

		engine.TakeState(string(c.Overleash.CachedJson()))

		ctx := createContextFromRequest(r)

		var apiResponse ApiResponse
		err := json.Unmarshal(engine.ResolveAll(ctx), &apiResponse)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return
		}

		result := frontendFromYggdrasil(apiResponse.Value, false)

		// Print result
		resultJson, _ := json.MarshalIndent(result, "", "  ")

		w.Write(resultJson)
	})

	s.HandleFunc("GET /api/frontend/all", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)

		engine.TakeState(string(c.Overleash.CachedJson()))

		ctx := createContextFromRequest(r)

		var apiResponse ApiResponse
		err := json.Unmarshal(engine.ResolveAll(ctx), &apiResponse)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return
		}

		result := frontendFromYggdrasil(apiResponse.Value, true)

		// Print result
		resultJson, _ := json.MarshalIndent(result, "", "  ")

		w.Write(resultJson)
	})

	s.HandleFunc("POST /api/frontend/client/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	s.HandleFunc("POST /api/frontend/client/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
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
	Name           string         `json:"name"`
	Enabled        bool           `json:"enabled"`
	Payload        VariantPayload `json:"payload"`
	FeatureEnabled bool           `json:"feature_enabled"`
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
	var toggles []EvaluatedToggle

	for name, resolved := range res {
		if includeAll || resolved.Enabled {
			toggle := EvaluatedToggle{
				Name:    name,
				Enabled: resolved.Enabled,
				Variant: EvaluatedVariant{
					Name:           resolved.Variant.Name,
					Enabled:        resolved.Variant.Enabled,
					Payload:        resolved.Variant.Payload,
					FeatureEnabled: resolved.Variant.FeatureEnabled,
				},
				ImpressionData: resolved.ImpressionData,
			}
			toggles = append(toggles, toggle)
		}
	}

	return FrontendResult{Toggles: toggles}
}

func createContextFromRequest(r *http.Request) *unleashengine.Context {
	return unleashengine.NewContext(
		getQuery(r, "userId"),
		getQuery(r, "sessionId"),
		getQuery(r, "environment"),
		getQuery(r, "appName"),
		getQuery(r, "currentTime"),
		getQuery(r, "remoteAddress"),
		nil,
	)
}

func getQuery(r *http.Request, name string) *string {
	if r.URL.Query().Has(name) {
		return nil
	}

	result := r.URL.Query().Get(name)

	return &result
}
