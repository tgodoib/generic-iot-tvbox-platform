//go:build ignore

// Template for secrets.go (not tracked by git).
// Copy to secrets.go, remove the build tag above and fill in the server domain and the device's mTLS credentials:
//
//	cp secrets.example.go secrets.go
package main

const serverAddr = "<server domain>"

const clientKey = "-----BEGIN EC PRIVATE KEY-----\n<device private key>\n-----END EC PRIVATE KEY-----"

const clientCrt = "-----BEGIN CERTIFICATE-----\n<device certificate>\n-----END CERTIFICATE-----"
