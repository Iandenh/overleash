package proxy

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

var (
	httpClient *http.Client
)

type Proxy struct {
	upstream string
}

func init() {
	httpClient = &http.Client{
		// Do not auto-follow redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 10 * time.Second,
	}
}

func New(upstream string) *Proxy {
	return &Proxy{upstream: upstream}
}

func (p *Proxy) ProxyRequest(w http.ResponseWriter, req *http.Request) error {
	newUrl, err := url.Parse(p.upstream)

	if err != nil {
		return err
	}

	newUrl.Path = path.Join(newUrl.Path, req.URL.Path)
	newQuery := newUrl.Query()

	for name, values := range req.URL.Query() {
		for _, value := range values {
			newQuery.Add(name, value)
		}
	}

	newUrl.RawQuery = newQuery.Encode()
	req.RequestURI = ""
	req.URL = newUrl
	req.Host = newUrl.Host

	proxiedResponse, err := httpClient.Do(req)

	if err != nil {
		return err
	}

	defer proxiedResponse.Body.Close()

	copyHeader(w.Header(), proxiedResponse.Header)
	w.WriteHeader(proxiedResponse.StatusCode)
	_, err = io.Copy(w, proxiedResponse.Body)

	if err != nil {
		return err
	}

	return nil
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		if dst.Get(k) != "" {
			continue
		}

		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
