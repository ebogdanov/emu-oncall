package grafana

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ebogdanov/emu-oncall/internal/metrics"
)

type apiClient struct {
	httpClient  *http.Client
	token       string
	promMetrics *metrics.Storage
}

type httpClient interface {
	Get(context.Context, string) ([]byte, error)
	Post(context.Context, string, string) ([]byte, error)
}

func newAPIClient(authToken string, storage *metrics.Storage) *apiClient {
	apiClient := &apiClient{
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
		token:       authToken,
		promMetrics: storage,
	}

	return apiClient
}

func (a *apiClient) Get(ctx context.Context, apiURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, apiURL, http.NoBody)

	if err != nil {
		return nil, err
	}

	return a.send(ctx, req)
}

func (a *apiClient) Post(ctx context.Context, apiURL, body string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer([]byte(body)))

	if err != nil {
		return nil, err
	}

	return a.send(ctx, req)
}

func (a *apiClient) send(ctx context.Context, req *http.Request) ([]byte, error) {
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "emu-oncall/0.0.1")

	if a.token != "" {
		req.Header.Add("Authorization", a.token)
	}

	code := 0
	startTime := time.Now()
	defer func(startTime time.Time, code *int) {
		duration := time.Since(startTime)

		a.promMetrics.APIResponseTime.
			WithLabelValues(req.URL.Path, req.Method, fmt.Sprintf("%d", *code)).
			Observe(duration.Seconds())
	}(startTime, &code)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	code = resp.StatusCode
	bodyBytes, err := io.ReadAll(resp.Body)
	defer func() {
		_ = resp.Body.Close()
	}()
	if err != nil {
		return nil, err
	}

	return bodyBytes, err
}
