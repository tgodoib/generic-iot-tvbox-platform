package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// datasourceClient builds an HTTP client from datasource instance settings.
// Uses mTLS when both clientCert and clientKey are configured.
func datasourceClient(settings *backend.DataSourceInstanceSettings) (*http.Client, error) {
	cert := settings.DecryptedSecureJSONData["clientCert"]
	key := settings.DecryptedSecureJSONData["clientKey"]
	if cert != "" && key != "" {
		return newMTLSClient(cert, key)
	}
	return plainClient(), nil
}

// checkDatasourceHealth tests connectivity to the configured server.
func checkDatasourceHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	settings := req.PluginContext.DataSourceInstanceSettings
	client, err := datasourceClient(settings)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("failed to create TLS client: %s", err),
		}, nil
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, settings.URL, nil)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("invalid server URL: %s", err),
		}, nil
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("connection failed: %s", err),
		}, nil
	}
	defer resp.Body.Close()

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Connected successfully",
	}, nil
}

// handleDatasourceCallResource proxies requests from the frontend datasource
// to the external IoT server using an mTLS client.
func handleDatasourceCallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	settings := req.PluginContext.DataSourceInstanceSettings
	client, err := datasourceClient(settings)
	if err != nil {
		return sender.Send(jsonErrResponse(http.StatusInternalServerError, err.Error()))
	}

	baseURL := strings.TrimRight(settings.URL, "/")

	targetURL, ok := resolveTargetURL(baseURL, req.Method, req.Path)
	if !ok {
		return sender.Send(jsonErrResponse(http.StatusNotFound, "unknown resource path"))
	}

	var bodyReader io.Reader
	if len(req.Body) > 0 {
		bodyReader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, bodyReader)
	if err != nil {
		return sender.Send(jsonErrResponse(http.StatusInternalServerError, err.Error()))
	}

	for k, vals := range req.Headers {
		for _, v := range vals {
			httpReq.Header.Add(k, v)
		}
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return sender.Send(jsonErrResponse(http.StatusBadGateway, err.Error()))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return sender.Send(jsonErrResponse(http.StatusInternalServerError, err.Error()))
	}

	headers := make(map[string][]string)
	for k, vals := range resp.Header {
		headers[k] = vals
	}

	return sender.Send(&backend.CallResourceResponse{
		Status:  resp.StatusCode,
		Headers: headers,
		Body:    body,
	})
}

func jsonErrResponse(status int, msg string) *backend.CallResourceResponse {
	body, _ := json.Marshal(map[string]string{"message": msg})
	return &backend.CallResourceResponse{
		Status:  status,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    body,
	}
}

// resolveTargetURL maps frontend resource paths to external API URLs.
//
// Frontend path → external API
//   GET  health          → GET  {base}/          (connectivity probe)
//   GET  firmwares       → GET  {base}/firmware
//   POST firmwares       → POST {base}/firmware  (multipart: file, version, device_class)
//   POST update/{id}     → POST {base}/update/{id}  (JSON: type, value)
//   DELETE firmware/{id} → DELETE {base}/firmware/{id}
func resolveTargetURL(baseURL, method, path string) (string, bool) {
	switch {
	case method == http.MethodGet && path == "health":
		return baseURL + "/", true

	case method == http.MethodGet && path == "firmwares":
		return baseURL + "/firmware", true

	case method == http.MethodPost && path == "firmwares":
		return baseURL + "/firmware", true

	case method == http.MethodPost && strings.HasPrefix(path, "update/"):
		id := strings.TrimPrefix(path, "update/")
		return baseURL + "/firmware/update/" + id, true

	case method == http.MethodDelete && strings.HasPrefix(path, "firmware/"):
		id := strings.TrimPrefix(path, "firmware/")
		return baseURL + "/firmware/" + id, true
	}
	return "", false
}
