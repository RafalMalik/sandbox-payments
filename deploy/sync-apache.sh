#!/bin/bash
# Sync Apache vhost + TLS for payments.raspberry-vision.cloud
# Run on the apps VPS from the sandbox-payments tree: bash deploy/sync-apache.sh

set -euo pipefail

DOMAIN="${DOMAIN:-payments.raspberry-vision.cloud}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CERT="/etc/letsencrypt/live/${DOMAIN}/fullchain.pem"
HTTP_BOOTSTRAP="${ROOT_DIR}/deploy/apache/${DOMAIN}-bootstrap.conf"
HTTP_REDIRECT="${ROOT_DIR}/deploy/apache/${DOMAIN}.conf"
SSL_CONF="${ROOT_DIR}/deploy/apache/${DOMAIN}-le-ssl.conf"

cd "${ROOT_DIR}"

for f in "${HTTP_BOOTSTRAP}" "${HTTP_REDIRECT}" "${SSL_CONF}"; do
    if [[ ! -f "${f}" ]]; then
        echo "Missing conf: ${f}" >&2
        exit 1
    fi
done

sudo a2enmod ssl proxy proxy_http headers rewrite >/dev/null

if ! sudo test -f "${CERT}"; then
    echo "No TLS cert yet — bootstrapping HTTP vhost for certbot..."
    sudo cp "${HTTP_BOOTSTRAP}" "/etc/apache2/sites-available/${DOMAIN}.conf"
    sudo a2ensite "${DOMAIN}.conf" >/dev/null
    sudo apache2ctl configtest
    sudo systemctl reload apache2

    if sudo certbot --apache \
        -d "${DOMAIN}" \
        --non-interactive \
        --agree-tos \
        --register-unsafely-without-email \
        --redirect; then
        echo "TLS certificate issued."
    else
        echo "Certbot failed (check DNS A record for ${DOMAIN}) — keeping HTTP-only vhost."
        exit 0
    fi
fi

sudo cp "${HTTP_REDIRECT}" "/etc/apache2/sites-available/${DOMAIN}.conf"
sudo cp "${SSL_CONF}" "/etc/apache2/sites-available/${DOMAIN}-le-ssl.conf"
sudo a2ensite "${DOMAIN}.conf" >/dev/null
sudo a2ensite "${DOMAIN}-le-ssl.conf" >/dev/null

sudo apache2ctl configtest
sudo systemctl reload apache2

echo "Apache synced for ${DOMAIN}"
