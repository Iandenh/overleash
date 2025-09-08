package overleash

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Iandenh/overleash/internal/version"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/launchdarkly/eventsource"
)

const (
	unleashClientSpecHeader   string = "Unleash-Client-Spec"
	unleashIntervalHeader     string = "Unleash-Interval"
	unleashConnectionIdHeader string = "Unleash-Connection-Id"
	unleashAppNameHeader      string = "Unleash-Appname"
	unleashSdkHeader          string = "Unleash-Sdk"

	supportedSpecVersion string = "5.1.0"
)

type client interface {
	getFeatures(token string) (*FeatureFile, error)
	validateToken(token string) (*EdgeToken, error)
	registerClient(token *EdgeToken) error
	bulkMetrics(token string, applications []*ClientData, metrics []*MetricsData) error
	streamFeatures(token string, channel chan eventsource.Event) error
}

type overleashClient struct {
	ctx          context.Context
	upstream     string
	httpClient   *http.Client
	connectionId string
	interval     int
}

func newClient(upstream string, interval time.Duration, ctx context.Context) *overleashClient {
	return &overleashClient{
		upstream:     upstream,
		interval:     int(interval.Seconds()),
		connectionId: uuid.New().String(),
		httpClient: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{},
			},
		},
		ctx: ctx,
	}
}

type validationRequest struct {
	Tokens []string `json:"tokens"`
}

type registerRequest struct {
	AppName      string    `json:"appName"`
	InstanceId   string    `json:"instanceId"`
	ConnectionId string    `json:"connectionId"`
	SdkVersion   string    `json:"sdkVersion"`
	Strategies   []string  `json:"strategies"`
	Started      time.Time `json:"started"`
	Interval     int       `json:"interval"`
	Environment  string    `json:"environment"`
}

type metricEnv struct {
	FeatureName string           `json:"featureName"`
	AppName     string           `json:"appName"`
	Environment string           `json:"environment"`
	Timestamp   time.Time        `json:"timestamp"`
	Yes         int64            `json:"yes"`
	No          int64            `json:"no"`
	Variants    map[string]int32 `json:"variants"`
}

func fromMetricData(data []*MetricsData) []*metricEnv {
	metrics := make([]*metricEnv, 0)

	for _, m := range data {
		for f, toggle := range m.Bucket.Toggles {
			variants := toggle.Variants

			if variants == nil {
				variants = make(map[string]int32)
			}

			metrics = append(metrics, &metricEnv{
				FeatureName: f,
				AppName:     m.AppName,
				Environment: m.Environment,
				Timestamp:   m.Bucket.Start,
				Yes:         int64(toggle.Yes),
				No:          int64(toggle.No),
				Variants:    variants,
			})
		}
	}

	return metrics
}

type clientEnv struct {
	ConnectVia   ConnectVia `json:"connectVia"`
	AppName      string     `json:"appName"`
	InstanceID   string     `json:"instanceId"`
	ConnectionId string     `json:"connectionId"`
	Environment  string     `json:"environment"`
	SDKVersion   string     `json:"sdkVersion"`
	Strategies   []string   `json:"strategies"`
	Started      time.Time  `json:"started"`
	Interval     int64      `json:"interval"`
}

func fromClientData(data []*ClientData, via ConnectVia) []*clientEnv {
	metrics := make([]*clientEnv, 0)

	for _, m := range data {
		metrics = append(metrics, &clientEnv{
			ConnectVia:   via,
			AppName:      m.AppName,
			InstanceID:   m.InstanceID,
			ConnectionId: m.ConnectionId,
			Environment:  m.Environment,
			SDKVersion:   m.SDKVersion,
			Strategies:   m.Strategies,
			Started:      m.Started,
			Interval:     m.Interval,
		})
	}

	return metrics
}

type validationResponse struct {
	Tokens []*EdgeToken `json:"tokens"`
}

func (c *overleashClient) getFeatures(token string) (*FeatureFile, error) {
	req, err := http.NewRequest(http.MethodGet, c.upstream+"/api/client/features", nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", token)
	req.Header.Add(unleashClientSpecHeader, supportedSpecVersion)
	req.Header.Add(unleashAppNameHeader, "Overleash")
	req.Header.Add(unleashConnectionIdHeader, c.connectionId)
	req.Header.Add(unleashIntervalHeader, strconv.Itoa(c.interval))
	req.Header.Add(unleashSdkHeader, "overleash@"+version.Version)

	res, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	response, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	features := &FeatureFile{}

	err = json.Unmarshal(response, features)

	if err != nil {
		return nil, err
	}

	return features, nil
}

func (c *overleashClient) validateToken(token string) (*EdgeToken, error) {
	req, err := http.NewRequest(http.MethodPost, c.upstream+"/edge/validate", nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(unleashClientSpecHeader, supportedSpecVersion)
	req.Header.Add(unleashAppNameHeader, "Overleash")
	req.Header.Add(unleashConnectionIdHeader, c.connectionId)
	req.Header.Add(unleashIntervalHeader, strconv.Itoa(c.interval))
	req.Header.Add(unleashSdkHeader, "overleash@"+version.Version)

	requestData := validationRequest{Tokens: []string{token}}
	tokenJson, err := json.Marshal(requestData)

	if err != nil {
		return nil, err
	}

	req.Body = io.NopCloser(bytes.NewReader(tokenJson))

	res, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	response, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	result := &validationResponse{}

	err = json.Unmarshal(response, result)

	if err != nil {
		return nil, err
	}

	tokens := result.Tokens

	if len(tokens) == 0 {
		return nil, errors.New("No tokens found")
	}

	return tokens[0], nil
}

func (c *overleashClient) registerClient(token *EdgeToken) error {
	req, err := http.NewRequest(http.MethodPost, c.upstream+"/api/client/register", nil)

	if err != nil {
		return err
	}

	overleashVersion := "overleash@" + version.Version
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token.Token)
	req.Header.Add(unleashClientSpecHeader, supportedSpecVersion)
	req.Header.Add(unleashAppNameHeader, "Overleash")
	req.Header.Add(unleashConnectionIdHeader, c.connectionId)
	req.Header.Add(unleashIntervalHeader, strconv.Itoa(c.interval))
	req.Header.Add(unleashSdkHeader, overleashVersion)

	requestData := registerRequest{
		AppName:      "Overleash",
		InstanceId:   c.connectionId,
		ConnectionId: c.connectionId,
		SdkVersion:   overleashVersion,
		Strategies:   make([]string, 0),
		Started:      time.Now(),
		Interval:     c.interval,
		Environment:  token.Environment,
	}

	requestJson, err := json.Marshal(requestData)

	if err != nil {
		return err
	}

	req.Body = io.NopCloser(bytes.NewReader(requestJson))

	res, err := c.httpClient.Do(req)

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d", res.StatusCode)
	}

	return nil
}

func (c *overleashClient) bulkMetrics(token string, applications []*ClientData, metrics []*MetricsData) error {
	req, err := http.NewRequest(http.MethodPost, c.upstream+"/api/client/metrics/bulk", nil)
	if err != nil {
		return err
	}

	overleashVersion := "overleash@" + version.Version
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)
	req.Header.Add(unleashClientSpecHeader, supportedSpecVersion)
	req.Header.Add(unleashAppNameHeader, "Overleash")
	req.Header.Add(unleashConnectionIdHeader, c.connectionId)
	req.Header.Add(unleashIntervalHeader, strconv.Itoa(c.interval))
	req.Header.Add(unleashSdkHeader, overleashVersion)

	via := ConnectVia{
		AppName:    "Overleash",
		InstanceID: c.connectionId,
	}

	requestData := struct {
		Applications []*clientEnv `json:"applications"`
		Metrics      []*metricEnv `json:"metrics"`
	}{
		Applications: fromClientData(applications, via),
		Metrics:      fromMetricData(metrics),
	}

	requestJson, err := json.Marshal(requestData)

	if err != nil {
		return err
	}

	req.Body = io.NopCloser(bytes.NewReader(requestJson))

	res, err := c.httpClient.Do(req)

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("invalid status code: %d", res.StatusCode)
	}

	return nil
}

func (c *overleashClient) streamFeatures(token string, channel chan eventsource.Event) error {
	req, err := http.NewRequest(http.MethodGet, c.upstream+"/api/client/streaming", nil)

	if err != nil {
		return err
	}

	overleashVersion := "overleash@" + version.Version
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)
	req.Header.Add("X-Overleash", "yes")
	req.Header.Add(unleashClientSpecHeader, supportedSpecVersion)
	req.Header.Add(unleashAppNameHeader, "Overleash")
	req.Header.Add(unleashConnectionIdHeader, c.connectionId)
	req.Header.Add(unleashIntervalHeader, strconv.Itoa(c.interval))
	req.Header.Add(unleashSdkHeader, overleashVersion)

	stream, err := eventsource.SubscribeWithRequestAndOptions(req,
		eventsource.StreamOptionCanRetryFirstConnection(-time.Second*3),
		eventsource.StreamOptionUseBackoff(5*time.Minute),
		eventsource.StreamOptionUseJitter(0.5),
		eventsource.StreamOptionErrorHandler(func(err error) eventsource.StreamErrorHandlerResult {
			return eventsource.StreamErrorHandlerResult{CloseNow: false}
		}),
	)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("SSE subscription panic recovered: %v", r)
				log.Error(err.Error())
			}
		}()

		for {
			select {
			case event := <-stream.Events:
				if event != nil {
					channel <- event
				}
			case <-c.ctx.Done():
				stream.Close()
				return
			}
		}
	}()

	return nil
}
