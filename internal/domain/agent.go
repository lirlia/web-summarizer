package domain

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/m-mizutani/goerr"
)

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
const redirectCount = 3

type httpAgent struct {
	client  *http.Client
	headers map[string]string
}

type Agent interface {
	Get(ctx context.Context, url string) ([]byte, error)
	Post(ctx context.Context, url string, body []byte) ([]byte, error)
}

func NewHTTPAgent(opts ...AgentOption) Agent {

	client := &http.Client{
		Transport: &http.Transport{
			// SSL証明書が無いサイトでも取得できるようにする
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	a := &httpAgent{
		client:  client,
		headers: map[string]string{},
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

type AgentOption func(*httpAgent)

func WithHeaders(headers map[string]string) AgentOption {
	return func(a *httpAgent) {
		a.headers = headers
	}
}

func (a *httpAgent) Post(_ context.Context, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, goerr.Wrap(err)
	}

	req.Header.Set("User-Agent", userAgent)
	for k, v := range a.headers {
		req.Header.Set(k, v)
	}

	res, err := a.client.Do(req)
	if err != nil {
		return nil, goerr.Wrap(err)
	}
	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, goerr.Wrap(err)
	}

	return buf, nil
}

func (a *httpAgent) Get(_ context.Context, url string) ([]byte, error) {
	return a.get(url, redirectCount)
}

func (a *httpAgent) get(url string, cnt int) ([]byte, error) {
	if cnt <= 0 {
		return nil, goerr.New("too many redirects")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, goerr.Wrap(err)
	}

	// for some sites, you need to set a user agent
	req.Header.Set("User-Agent", userAgent)
	for k, v := range a.headers {
		req.Header.Set(k, v)
	}

	res, err := a.client.Do(req)
	if err != nil {
		return nil, goerr.Wrap(err)
	}
	defer res.Body.Close()

	// Handle redirects
	if res.StatusCode >= 300 && res.StatusCode < 400 {
		redirectURL := res.Header.Get("Location")
		return a.get(redirectURL, cnt-1)
	}

	// Handle errors
	if res.StatusCode != 200 {
		return nil, goerr.New(fmt.Sprintf("HTTP status code is not 200: %d", res.StatusCode))
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, goerr.Wrap(err)
	}

	return buf, nil
}
