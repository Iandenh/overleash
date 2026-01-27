package tests

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Iandenh/overleash/config"
	"github.com/Iandenh/overleash/overleash"
	"github.com/Iandenh/overleash/server"
	"github.com/Iandenh/overleash/unleashengine"
	"github.com/Unleash/unleash-go-sdk/v5"
	unleashContext "github.com/Unleash/unleash-go-sdk/v5/context"
)

const (
	specFolder = "./client-specification/specifications"
	serverPort = ":1244"
	serverURL  = "http://localhost" + serverPort
)

var specIndex = filepath.Join(specFolder, "index.json")

type TestCase struct {
	Description    string                `json:"description"`
	Context        unleashengine.Context `json:"context"`
	ToggleName     string                `json:"toggleName"`
	ExpectedResult bool                  `json:"expectedResult"`
}

type expectedVariantResult struct {
	overleash.Variant
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

type PhpResult struct {
	Result     bool               `json:"result"`
	ToggleName string             `json:"toggleName"`
	Variant    *overleash.Variant `json:"variant"`
}

// --- Tests ---

// TestFrontendApi is the original "TestShit".
// It tests the API using the Go HTTP Client.
func TestFrontendApi(t *testing.T) {
	// 1. Start Server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	o := startServer(t, ctx)

	// Client setup
	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{},
		},
		Timeout: 10 * time.Second,
	}

	waitForServer(t, httpClient, serverURL+"/health")

	// 2. Get Definitions
	definitions, err := getDefinitions()
	if err != nil {
		t.SkipNow()
	}

	// 3. Run Tests
	for _, definition := range definitions {
		t.Run(definition.Name, func(t *testing.T) {
			o.LoadFeatureFile(definition.State)

			if len(definition.Tests) == 0 && len(definition.VariantTests) == 0 {
				t.Skip("No standard tests in this definition")
			}

			for _, testCase := range definition.Tests {
				cont, err := json.Marshal(testCase.Context)
				if err != nil {
					t.Fatal(err)
				}

				req, err := http.NewRequest(http.MethodPost, serverURL+"/api/frontend/features/"+testCase.ToggleName, io.NopCloser(bytes.NewReader(cont)))
				if err != nil {
					t.Fatal(err)
				}

				res, err := httpClient.Do(req)
				if err != nil {
					t.Fatal(err)
				}
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

			for _, testCase := range definition.VariantTests {
				cont, err := json.Marshal(testCase.Context)
				if err != nil {
					t.Fatal(err)
				}

				req, err := http.NewRequest(http.MethodPost, serverURL+"/api/frontend/features/"+testCase.ToggleName, io.NopCloser(bytes.NewReader(cont)))
				if err != nil {
					t.Fatal(err)
				}

				res, err := httpClient.Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer res.Body.Close()

				response, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}

				var j struct {
					Enabled bool               `json:"enabled"`
					Variant *overleash.Variant `json:"variant"`
				}
				json.Unmarshal(response, &j)

				if reflect.DeepEqual(testCase.ExpectedResult.Variant, j.Variant) {
					t.Fatalf("%s: expected %v, got %v", testCase.Description, testCase.ExpectedResult.Name, j.Variant.Name)
				}
			}
		})
	}
}

// TestPhpIntegration runs the PHP docker container against the Go Server.
func TestPhpIntegration(t *testing.T) {
	// 1. Start Server (New instance for this test)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	o := startServer(t, ctx)

	// Check if Docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not found, skipping PHP integration tests")
	}

	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{},
		},
		Timeout: 10 * time.Second,
	}
	waitForServer(t, httpClient, serverURL+"/health")

	// 2. Get Definitions
	definitions, err := getDefinitions()
	if err != nil {
		t.Fatalf("Failed to load definitions: %v", err)
	}

	// 3. Run Tests
	for _, definition := range definitions {
		t.Run(definition.Name, func(t *testing.T) {
			o.LoadFeatureFile(definition.State)

			if len(definition.Tests) == 0 && len(definition.VariantTests) == 0 {
				t.Skip("No standard tests in this definition")
			}

			// Execute PHP Container
			phpResults, err := runPHPInDocker(defToJSON(t, definition))
			if err != nil {
				t.Fatalf("PHP execution failed: %v", err)
			}

			// Verify Results
			for _, testCase := range definition.Tests {
				phpRes, exists := phpResults[testCase.Description]
				if !exists {
					t.Errorf("Test case '%s' missing from PHP output", testCase.Description)
					continue
				}
				if phpRes.Result != testCase.ExpectedResult {
					t.Errorf("FAIL: %s\n\tToggle: %s\n\tExpected: %v\n\tGot (PHP): %v",
						testCase.Description, testCase.ToggleName, testCase.ExpectedResult, phpRes.Result)
				}
			}

			for _, testCase := range definition.VariantTests {
				phpRes, exists := phpResults[testCase.Description]
				if !exists {
					t.Errorf("Test case '%s' missing from PHP output", testCase.Description)
					continue
				}

				if reflect.DeepEqual(testCase.ExpectedResult.Variant, phpRes.Variant) {
					t.Fatalf("%s: expected %v, got %v", testCase.Description, testCase.ExpectedResult.Name, phpRes.Variant.Name)
				}
			}
		})
	}
}

func TestGoIntegration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	o := startServer(t, ctx)

	httpClient := &http.Client{Timeout: 2 * time.Second}
	waitForServer(t, httpClient, serverURL+"/health")

	definitions, err := getDefinitions()
	if err != nil {
		t.Fatalf("Failed to load definitions: %v", err)
	}

	options := []unleash.ConfigOption{
		unleash.WithAppName("Test"),
		unleash.WithUrl(serverURL + "/api"),
	}

	for _, definition := range definitions {

		if definition.Name == "16-strategy-variants" {
			t.Skip("Skipped for now. Flaky in Go unleash client")
		}

		t.Run(definition.Name, func(t *testing.T) {
			o.LoadFeatureFile(definition.State)

			if len(definition.Tests) == 0 && len(definition.VariantTests) == 0 {
				t.Skip("No standard tests in this definition")
			}

			goClient, err := unleash.NewClient(options...)

			if err != nil {
				t.Fatal(err)
			}

			goClient.WaitForReady()

			for _, testCase := range definition.Tests {
				result := goClient.IsEnabled(testCase.ToggleName, unleash.WithContext(toUnleashContext(&testCase.Context)), unleash.WithFallback(false))

				if result != testCase.ExpectedResult {
					t.Errorf("FAIL: %s\n\tToggle: %s\n\tExpected: %v\n\tGot (PHP): %v",
						testCase.Description, testCase.ToggleName, testCase.ExpectedResult, result)
				}
			}

			for _, testCase := range definition.VariantTests {
				result := goClient.GetVariant(testCase.ToggleName, unleash.WithVariantContext(toUnleashContext(&testCase.Context)))

				if result.Name != testCase.ExpectedResult.Variant.Name ||
					result.Payload.Type != testCase.ExpectedResult.Variant.Payload.Type ||
					result.FeatureEnabled != testCase.ExpectedResult.SpecFeatureEnabled {
					t.Fatalf("%s: expected %+v, got %+v", testCase.Description, testCase.ExpectedResult, result)
				}
			}
		})
	}
}

func toUnleashContext(ctx *unleashengine.Context) unleashContext.Context {
	if ctx == nil {
		return unleashContext.Context{}
	}

	return unleashContext.Context{
		UserId:        stringValue(ctx.UserId),
		SessionId:     stringValue(ctx.SessionId),
		RemoteAddress: stringValue(ctx.RemoteAddress),
		Environment:   stringValue(ctx.Environment),
		AppName:       stringValue(ctx.AppName),
		CurrentTime:   stringValue(ctx.CurrentTime),
		Properties:    ctx.Properties,
	}
}

func stringValue(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

// --- Helper Functions ---
func startServer(t *testing.T, ctx context.Context) *overleash.OverleashContext {
	o := overleash.NewOverleash(&config.Config{
		URL:               "",
		Upstream:          "",
		Token:             "default:development.unleash-insecure-api-token",
		ListenAddress:     serverPort,
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

	go func() {
		// We ignore the error here as it will return http.ErrServerClosed on shutdown
		server.New(o, ctx).Start()
	}()

	// Small sleep to allow socket bind
	time.Sleep(100 * time.Millisecond)
	return o
}

func waitForServer(t *testing.T, client *http.Client, url string) {
	for i := 0; i < 50; i++ {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("Server failed to start in time")
}

func getDefinitions() ([]TestDefinition, error) {
	index, err := os.Open(specIndex)
	if err != nil {
		return nil, err
	}
	defer index.Close()
	var testFiles []string
	if err := json.NewDecoder(index).Decode(&testFiles); err != nil {
		panic(err)
	}

	definitions := make([]TestDefinition, len(testFiles))
	for idx, testFile := range testFiles {
		test, err := os.Open(filepath.Join(specFolder, testFile))
		if err != nil {
			panic(err)
		}
		var testDef TestDefinition
		if err := json.NewDecoder(test).Decode(&testDef); err != nil {
			panic(err)
		}
		definitions[idx] = testDef
	}
	return definitions, nil
}

func defToJSON(t *testing.T, v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func runPHPInDocker(inputData []byte) (map[string]PhpResult, error) {
	cmd := exec.Command("docker", "run",
		"--rm",           // Cleanup container after run
		"--network=host", // Use host network to reach localhost:1244
		"-i",             // Interactive (reads from stdin)
		"-e", "UNLEASH_API_URL=http://localhost"+serverPort+"/api",
		"overleash-php-test", // Image Name
	)

	cmd.Stdin = bytes.NewReader(inputData)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("docker run error: %v | Stderr: %s", err, stderr.String())
	}

	results := make(map[string]PhpResult)
	if err := json.Unmarshal(out.Bytes(), &results); err != nil {
		return nil, fmt.Errorf("failed to parse PHP output: %v | Output: %s", err, out.String())
	}

	return results, nil
}
