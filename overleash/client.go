package overleash

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Iandenh/overleash/internal/version"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	unleashClientSpecHeader   string = "Unleash-Client-Spec"
	unleashIntervalHeader     string = "Unleash-Interval"
	unleashConnectionIdHeader string = "Unleash-Connection-Id"
	unleashAppNameHeader      string = "Unleash-Appname"
	unleashSdkHeader          string = "Unleash-Sdk"

	supportedSpecVersion string = "5.1.0"
)

type OverleashClient struct {
	url          string
	httpClient   *http.Client
	connectionId string
	interval     int
}

func NewClient(url string, interval int) *OverleashClient {
	return &OverleashClient{
		url:          url,
		interval:     interval * 60,
		connectionId: uuid.New().String(),
		httpClient: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

type validationRequest struct {
	Tokens []string `json:"tokens"`
}

type registerRequest struct {
	AppName     string    `json:"appName"`
	InstanceId  string    `json:"instanceId"`
	SdkVersion  string    `json:"sdkVersion"`
	Strategies  []string  `json:"strategies"`
	Started     time.Time `json:"started"`
	Interval    int       `json:"interval"`
	Environment string    `json:"environment"`
}

type validationResponse struct {
	Tokens []*EdgeToken `json:"tokens"`
}

func (c *OverleashClient) getFeatures(token string) (*FeatureFile, error) {
	req, err := http.NewRequest(http.MethodGet, c.url+"/api/client/features", nil)

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

func (c *OverleashClient) validateToken(token string) (*EdgeToken, error) {
	req, err := http.NewRequest(http.MethodPost, c.url+"/edge/validate", nil)

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
		return nil, errors.New("no tokens found")
	}

	return tokens[0], nil
}

func (c *OverleashClient) registerClient(token *EdgeToken) error {
	req, err := http.NewRequest(http.MethodPost, c.url+"/api/client/register", nil)

	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token.Token)
	req.Header.Add(unleashClientSpecHeader, supportedSpecVersion)
	req.Header.Add(unleashAppNameHeader, "Overleash")
	req.Header.Add(unleashConnectionIdHeader, c.connectionId)
	req.Header.Add(unleashIntervalHeader, strconv.Itoa(c.interval))
	req.Header.Add(unleashSdkHeader, "overleash@"+version.Version)

	requestData := registerRequest{
		AppName:     "Overleash",
		InstanceId:  "Overleash",
		SdkVersion:  "overleash@" + version.Version,
		Strategies:  make([]string, 0),
		Started:     time.Now(),
		Interval:    c.interval * 1000, // in milliseconds
		Environment: token.Environment,
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
