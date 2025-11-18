package tests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"testing"
	"time"

	"github.com/Iandenh/overleash/config"
	"github.com/Iandenh/overleash/overleash"
	"github.com/Iandenh/overleash/server"
	"github.com/Iandenh/overleash/unleashengine"
)

const specFolder = "./client-specification/specifications"

var specIndex = filepath.Join(specFolder, "index.json")

type TestCase struct {
	Description    string                `json:"description"`
	Context        unleashengine.Context `json:"context"`
	ToggleName     string                `json:"toggleName"`
	ExpectedResult bool                  `json:"expectedResult"`
}

type expectedVariantResult struct {
	overleash.Variant
	// SpecFeatureEnabled represents the spec's feature_enabled field which has a
	// different JSON field name than api.Variant
	SpecFeatureEnabled bool `json:"feature_enabled"`
}

type VariantTestCase struct {
	Description    string                 `json:"description"`
	Context        unleashengine.Context  `json:"context"`
	ToggleName     string                 `json:"toggleName"`
	ExpectedResult *expectedVariantResult `json:"expectedResult"`
}

type TestDefinition struct {
	Name         string                `json:"name"`
	State        overleash.FeatureFile `json:"state"`
	Tests        []TestCase            `json:"tests"`
	VariantTests []VariantTestCase     `json:"variantTests"`
}

func TestShit(t *testing.T) {
	ctx, cancel := signal.NotifyContext(t.Context(), os.Interrupt)
	defer cancel()

	o := overleash.NewOverleash(&config.Config{
		URL:               "",
		Upstream:          "",
		Token:             "default:development.unleash-insecure-api-token",
		ListenAddress:     ":1244",
		Reload:            "0",
		Verbose:           false,
		RegisterMetrics:   false,
		PrometheusMetrics: false,
		PrometheusPort:    0,
		Register:          false,
		Headless:          false,
		Streamer:          false,
		EnableFrontend:    true,
		Delta:             false,
		EnvFromToken:      false,
		Webhook:           false,
		Storage:           "null",
		RedisAddr:         "",
		RedisPassword:     "",
		RedisDB:           0,
		RedisChannel:      "",
		RedisSentinel:     false,
		RedisMaster:       "",
		RedisSentinels:    "",
	})

	httpClient := &http.Client{
		// Do not auto-follow redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{}, // Use system's default trusted CA pool
		},
		Timeout: 10 * time.Second,
	}

	go server.New(o, ctx).Start()

	time.Sleep(100 * time.Millisecond)

	definitions, err := definitions()

	if err != nil {
		t.SkipNow()
	}

	for _, definition := range definitions {
		o.LoadFeatureFile(definition.State)

		for _, testCase := range definition.Tests {
			cont, err := json.Marshal(testCase.Context)
			req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:1244/api/frontend/features/"+testCase.ToggleName, io.NopCloser(bytes.NewReader(cont)))

			if err != nil {
				t.Fatal(err)
			}

			res, err := httpClient.Do(req)

			defer res.Body.Close()

			response, err := io.ReadAll(res.Body)

			if err != nil {
				t.Fatal(err)
			}

			var j struct {
				Enabled bool `json:"enabled"`
			}

			json.Unmarshal(response, &j)

			if j.Enabled != testCase.ExpectedResult {
				t.Fatalf("%s: expected %v, got %v", testCase.Description, testCase.ExpectedResult, j.Enabled)
			}
		}
	}

	cancel()
}

func definitions() ([]TestDefinition, error) {
	index, err := os.Open(specIndex)
	if err != nil {
		return nil, err
	}

	defer index.Close()
	var testFiles []string
	dec := json.NewDecoder(index)
	err = dec.Decode(&testFiles)

	if err != nil {
		panic(err)
	}

	definitions := make([]TestDefinition, len(testFiles))

	for idx, testFile := range testFiles {
		test, err := os.Open(filepath.Join(specFolder, testFile))

		if err != nil {
			panic(err)
		}

		var testDef TestDefinition
		dec := json.NewDecoder(test)
		err = dec.Decode(&testDef)
		if err != nil {
			panic(err)
		}

		definitions[idx] = testDef
	}

	return definitions, nil
}
