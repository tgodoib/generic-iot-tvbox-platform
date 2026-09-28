#!/usr/bin/env bash
set -euo pipefail

CA="$1"
NAME="$2"

mkdir -p /root/certs/clients

# chave privada do cliente
openssl ecparam -name prime256v1 -genkey -noout \
  -out /root/certs/clients/${NAME}.key

# csr
openssl req -new \
  -key /root/certs/clients/${NAME}.key \
  -out /root/certs/clients/${NAME}.csr \
  -subj "/CN=${NAME}/O=SSAIOT"

# assinatura pela CA
openssl x509 -req -days 365 \
  -in /root/certs/clients/${NAME}.csr \
  -CA /root/certs/cas/${CA}.crt \
  -CAkey /root/certs/cas/${CA}.key \
  -CAcreateserial \
  -out /root/certs/clients/${NAME}.crt

# cleanup
rm /root/certs/clients/${NAME}.csr