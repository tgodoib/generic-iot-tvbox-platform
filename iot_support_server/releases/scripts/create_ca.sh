#!/usr/bin/env bash
set -euo pipefail

NAME="$1"

mkdir -p /root/certs/cas

# chave privada
openssl ecparam -name prime256v1 -genkey -noout \
  -out /root/certs/cas/${NAME}.key

# certificado da CA (10 anos)
openssl req -new -x509 -days 3650 \
  -key /root/certs/cas/${NAME}.key \
  -out /root/certs/cas/${NAME}.crt \
  -subj "/CN=${NAME}/O=SSAIOT"