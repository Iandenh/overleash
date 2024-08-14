package proxy

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
)

type Proxy struct {
	upstream string
}

func New(upstream string) *Proxy {
	return &Proxy{upstream: upstream}
}

func (p *Proxy) ProxyRequest(w http.ResponseWriter, req *http.Request) error {
	newUrl, _ := url.Parse(p.upstream)

	concatPath := req.URL.Path
	if concatPath != "" {
		newUrl.Path = newUrl.Path + concatPath
	}

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

	proxiedResponse, err := createProxyClient().Do(req)

	if err != nil {
		return err
	}

	copyHeader(w.Header(), proxiedResponse.Header)
	w.WriteHeader(proxiedResponse.StatusCode)
	io.Copy(w, proxiedResponse.Body)

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

func createProxyClient() *http.Client {
	return &http.Client{
		// Do not auto-follow redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
