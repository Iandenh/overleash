package overleash

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
)

var (
	httpClient      *http.Client
	httpClientMutex sync.Mutex
)

type validationRequest struct {
	Tokens []string `json:"tokens"`
}

type validationResponse struct {
	Tokens []*EdgeToken `json:"tokens"`
}

func getFeatures(url, token string) (*FeatureFile, error) {
	httpClient := createHTTPClient()

	req, err := http.NewRequest(http.MethodGet, url+"/api/client/features", nil)

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", token)

	if err != nil {
		return nil, err
	}

	res, err := httpClient.Do(req)

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

func validateToken(url string, token string) (*EdgeToken, error) {
	httpClient := createHTTPClient()

	req, err := http.NewRequest(http.MethodPost, url+"/edge/validate", nil)

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	requestData := validationRequest{Tokens: []string{token}}
	tokenJson, err := json.Marshal(requestData)

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	req.Body = io.NopCloser(bytes.NewReader(tokenJson))

	res, err := httpClient.Do(req)

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

func createHTTPClient() *http.Client {
	httpClientMutex.Lock()
	defer httpClientMutex.Unlock()

	if httpClient == nil {
		httpClient = &http.Client{
			// Do not auto-follow redirects
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}
	return httpClient
}
