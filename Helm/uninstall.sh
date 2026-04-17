#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════════
#  HipsterShop — Helm Uninstall Script
#  Removes all HipsterShop Helm releases.
#  NOTE: App namespaces (hipster-backend, hipster-frontend, hipster-database)
#  have `helm.sh/resource-policy: keep` and are NOT deleted automatically.
#  To also delete namespaces, pass --delete-namespaces flag.
#
#  Usage:
#    ./Helm/uninstall.sh                        # uninstall, keep namespaces
#    ./Helm/uninstall.sh hipster                # custom Helm release namespace
#    ./Helm/uninstall.sh hipster --delete-namespaces
# ═══════════════════════════════════════════════════════════════════════════
set -euo pipefail

NS="${1:-hipster}"
DELETE_NS=false
[[ "${2:-}" == "--delete-namespaces" ]] && DELETE_NS=true

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; BOLD='\033[1m'; RESET='\033[0m'
info()    { echo -e "\033[0;36m${BOLD}[INFO]${RESET}  $*"; }
success() { echo -e "${GREEN}${BOLD}[ OK ]${RESET}  $*"; }
warn()    { echo -e "${YELLOW}${BOLD}[WARN]${RESET}  $*"; }

_uninstall() {
  local name="$1"
  if helm status "$name" -n "$NS" &>/dev/null; then
    info "Uninstalling $name"
    helm uninstall "$name" -n "$NS"
    success "$name removed"
  else
    warn "$name not found in namespace $NS — skipping"
  fi
}

echo -e "\n${BOLD}HipsterShop Helm Uninstall${RESET}  [namespace: ${YELLOW}$NS${RESET}]\n"

# Uninstall in reverse dependency order
_uninstall loadgenerator
_uninstall adservice
_uninstall assistantservice
_uninstall authservice
_uninstall cartservice
_uninstall checkoutservice
_uninstall currencyservice
_uninstall emailservice
_uninstall paymentservice
_uninstall productcatalogservice
_uninstall recommendationservice
_uninstall shippingservice
_uninstall gateway
_uninstall frontend
_uninstall mongodb
_uninstall backend-common

if $DELETE_NS; then
  echo -e "\n${RED}${BOLD}Deleting app namespaces...${RESET}"
  for ns in hipster-backend hipster-frontend hipster-database; do
    kubectl delete namespace "$ns" --ignore-not-found=true
    success "Namespace $ns deleted"
  done
fi

echo -e "\n${GREEN}${BOLD}✓ Uninstall complete${RESET}"
echo -e "  Helm namespace '$NS' still exists. Remove manually if desired:"
echo -e "  kubectl delete namespace $NS"
