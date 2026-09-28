#!/usr/bin/env bash

echo "== installing secrets =="
if [ -f ./config/ssaiot.env ]; then
    sudo install -d -m 700 /etc/ssaiot
    sudo install -m 600 ./config/ssaiot.env /etc/ssaiot/ssaiot.env
elif [ ! -f /etc/ssaiot/ssaiot.env ]; then
    echo "⚠️  config/ssaiot.env not found and /etc/ssaiot/ssaiot.env missing; aborting (see config/ssaiot.env.example)"
    exit 1
else
    echo "config/ssaiot.env not found, keeping existing /etc/ssaiot/ssaiot.env"
fi

echo "== installing nginx =="
if ! dpkg -s nginx >/dev/null 2>&1; then
    sudo apt update
    sudo apt install -y nginx
else
    echo "nginx already installed, skipping"
fi
if ! dpkg -s libnginx-mod-stream >/dev/null 2>&1; then
    sudo apt update
    sudo apt install -y libnginx-mod-stream
else
    echo "libnginx-mod-stream already installed, skipping"
fi
if ! command -v envsubst >/dev/null 2>&1; then
    sudo apt update
    sudo apt install -y gettext-base
fi
# Only our variables are substituted; nginx's own $vars are left untouched
sudo bash -c 'set -a; . /etc/ssaiot/ssaiot.env; set +a; envsubst "\${SERVER_DOMAIN} \${MINIO_ENDPOINT}"' \
    < ./config/nginx.conf.template | sudo tee /etc/nginx/nginx.conf >/dev/null
sudo nginx -t || { echo "⚠️  nginx config invalid, check SERVER_DOMAIN / MINIO_ENDPOINT"; exit 1; }
sudo systemctl restart nginx.service

echo "== installing mqtt broker =="
if ! dpkg -s mosquitto >/dev/null 2>&1; then
    sudo apt update
    sudo apt install -y mosquitto
else
    echo "mosquitto already installed, skipping"
fi
sudo cp ./config/mosquitto.conf /etc/mosquitto/conf.d/general.conf
sudo systemctl restart mosquitto.service

echo "== installing scripts =="
sudo mkdir -p /root/certs/scripts
sudo cp ./scripts/* /root/certs/scripts/
sudo chmod +x /root/certs/scripts/*

echo "== installing modules =="
sudo install -d /opt/ssaiot/bin
sudo install -m 755 ./bin/* /opt/ssaiot/bin/

echo "== installing services =="
for svc in ./systemd/*.service; do
    name=$(basename "$svc")
    sudo cp "$svc" "/etc/systemd/system/$name"
done

sudo systemctl daemon-reload

for svc in ./systemd/*.service; do
    name=$(basename "$svc")
    sudo systemctl enable "$name"
    sudo systemctl restart "$name" || echo "⚠️  Failed to restart $name"
done

echo "== adding ssaiot-status alias =="

ALIAS_FILE="$HOME/.bash_aliases"
ALIAS_NAME="ssaiot-status"

touch "$ALIAS_FILE"

if grep -q "$ALIAS_NAME" "$ALIAS_FILE"; then
  echo "Alias '$ALIAS_NAME' already exists in $ALIAS_FILE, nothing changed."
else
  cat >> "$ALIAS_FILE" << 'EOF'

alias ssaiot-status='for s in health_checker log_collector notification_system ota_server sensor_data_collector nginx mosquitto; do
  status=$(systemctl is-active ${s}.service)
  [ "$status" = "active" ] && echo "✅ $s: $status" || echo "❌ $s: $status"
done'
EOF
  echo "Alias '$ALIAS_NAME' added in $ALIAS_FILE"
fi

echo "== done =="