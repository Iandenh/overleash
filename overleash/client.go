package overleash

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

var (
	httpClient      *http.Client
	httpClientMutex sync.Mutex
)

func getFeatures(url, token string) (*FeatureFile, error) {
	httpClient := createHTTPClient()

	req, err := http.NewRequest(http.MethodGet, url, nil)

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
