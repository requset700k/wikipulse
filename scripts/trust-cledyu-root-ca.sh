#!/usr/bin/env bash
set -euo pipefail

namespace="cert-manager"
secret_name="cledyu-root-ca"
cert_path=""

usage() {
  cat << 'USAGE'
Usage:
  scripts/trust-cledyu-root-ca.sh [--cert-path PATH] [--namespace NAME] [--secret-name NAME]

Installs the Cledyu root CA public certificate into the local OS trust store.
If --cert-path is omitted, the public certificate is extracted from Kubernetes.

This script never reads or installs tls.key.
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --cert-path)
      cert_path="$2"
      shift 2
      ;;
    --namespace)
      namespace="$2"
      shift 2
      ;;
    --secret-name)
      secret_name="$2"
      shift 2
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

decode_base64() {
  if base64 --help 2>&1 | grep -q -- '--decode'; then
    base64 --decode
  else
    base64 -D
  fi
}

fetch_certificate() {
  if [[ -n "$cert_path" ]]; then
    return
  fi

  cert_path="$(mktemp "${TMPDIR:-/tmp}/cledyu-root-ca.XXXXXX.crt")"
  kubectl -n "$namespace" get secret "$secret_name" -o 'jsonpath={.data.tls\.crt}' |
    decode_base64 > "$cert_path"
}

install_macos() {
  sudo security add-trusted-cert \
    -d \
    -r trustRoot \
    -k /Library/Keychains/System.keychain \
    "$cert_path"
}

install_linux() {
  if command -v update-ca-certificates > /dev/null 2>&1; then
    sudo install -m 0644 "$cert_path" /usr/local/share/ca-certificates/cledyu-root-ca.crt
    sudo update-ca-certificates
    return
  fi

  if command -v update-ca-trust > /dev/null 2>&1; then
    sudo install -m 0644 "$cert_path" /etc/pki/ca-trust/source/anchors/cledyu-root-ca.crt
    sudo update-ca-trust
    return
  fi

  echo "Unsupported Linux trust store. Install the certificate manually: $cert_path" >&2
  exit 1
}

fetch_certificate

case "$(uname -s)" in
  Darwin)
    install_macos
    ;;
  Linux)
    install_linux
    ;;
  *)
    echo "Unsupported OS: $(uname -s)" >&2
    exit 1
    ;;
esac

echo "Installed Cledyu Root CA from: $cert_path"
echo "Restart the browser, then open https://keycloak.cledyu.local/admin/"
