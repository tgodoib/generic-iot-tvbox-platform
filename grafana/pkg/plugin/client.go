package plugin

import (
	"crypto/tls"
	"net/http"
	"time"
)

func newMTLSClient(certPEM, keyPEM string) (*http.Client, error) {
	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
				// System CA pool used by default — covers Let's Encrypt / public CAs.
			},
		},
		Timeout: 30 * time.Second,
	}, nil
}

func plainClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}
