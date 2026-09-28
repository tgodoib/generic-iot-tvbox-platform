package adapters

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"ssaiot_sim/ssaiot"
)

type Web struct {
	cfg *ssaiot.Config
}

func NewWeb(cfg *ssaiot.Config) *Web {
	return &Web{
		cfg: cfg,
	}
}

func (w *Web) getClient() (*http.Client, error) {
	cert, err := tls.X509KeyPair([]byte(w.cfg.ClientCrt), []byte(w.cfg.ClientKey))
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
	}, nil
}

// Get returns the response body as a stream; the caller must close it.
func (w *Web) Get(url string) (io.ReadCloser, error) {
	client, err := w.getClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
