#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════════
#  HipsterShop — Helm Install Script
#  Installs every microservice as its own Helm release so that
#  `helm list -n <NS>` shows each service by name.
#
#  Usage:
#    chmod +x ./Helm/install.sh
#    ./Helm/install.sh              # installs into namespace 'hipster'
#    ./Helm/install.sh my-namespace # installs into a custom namespace
#
#  Works with: bash, zsh, Git Bash (Windows), WSL
# ═══════════════════════════════════════════════════════════════════════════
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VALUES="$SCRIPT_DIR/values.yaml"
CHARTS="$SCRIPT_DIR/charts"
NS="${1:-hipster}"

# ── Colour helpers ──────────────────────────────────────────────────────────
GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; BOLD='\033[1m'; RESET='\033[0m'
info()    { echo -e "${CYAN}${BOLD}[INFO]${RESET}  $*"; }
success() { echo -e "${GREEN}${BOLD}[ OK ]${RESET}  $*"; }
section() { echo -e "\n${YELLOW}${BOLD}━━━ $* ━━━${RESET}"; }

# ── Helper: install or upgrade a single chart ───────────────────────────────
_install() {
  local name="$1"
  local chart="$CHARTS/$2"
  info "Installing ${BOLD}$name${RESET} from $chart"
  helm upgrade --install "$name" "$chart" \
    -f "$VALUES" \
    -n "$NS" \
    --create-namespace
  success "$name"
}

echo -e "\n${BOLD}HipsterShop Helm Install${RESET}"
echo -e "  Namespace : ${CYAN}$NS${RESET}"
echo -e "  Values    : ${CYAN}$VALUES${RESET}\n"

# ── Step 1: Helm release namespace ─────────────────────────────────────────
section "Step 1/5 — Helm Release Namespace"
kubectl create namespace "$NS" --dry-run=client -o yaml | kubectl apply -f -
success "Namespace '$NS' ready"

# ── Step 2: Infrastructure (creates app namespaces + shared resources) ──────
# MUST run before any service chart — they deploy into hipster-backend /
# hipster-frontend / hipster-database which these charts create.
section "Step 2/5 — Infrastructure"
_install backend-common backend-common   # → hipster-backend NS, ConfigMaps, Secrets, NetworkPolicy
_install mongodb         mongodb          # → hipster-database NS, StatefulSet, init/seed Jobs
_install frontend        frontend         # → hipster-frontend NS, Deployment, HPA, NetworkPolicy

# ── Step 3: Gateway (needs hipster-backend namespace) ───────────────────────
section "Step 3/5 — Gateway"
_install gateway gateway                 # → GatewayClass (if not exists), Gateway, HTTPRoutes

# ── Step 4: Backend microservices ───────────────────────────────────────────
section "Step 4/5 — Backend Microservices"
_install adservice            adservice
_install assistantservice     assistantservice
_install authservice          authservice
_install cartservice          cartservice
_install checkoutservice      checkoutservice
_install currencyservice      currencyservice
_install emailservice         emailservice
_install paymentservice       paymentservice
_install productcatalogservice productcatalogservice
_install recommendationservice recommendationservice
_install shippingservice      shippingservice

# ── Step 5: Traffic generator ───────────────────────────────────────────────
section "Step 5/5 — Load Generator"
_install loadgenerator loadgenerator     # → Deployment in hipster-frontend

# ── Summary ─────────────────────────────────────────────────────────────────
echo -e "\n${GREEN}${BOLD}🎉 All HipsterShop charts installed successfully!${RESET}"
echo -e "\nCheck release status:"
echo -e "  ${CYAN}helm list -n $NS${RESET}"
echo -e "\nCheck pods:"
echo -e "  ${CYAN}kubectl get pods -n hipster-backend${RESET}"
echo -e "  ${CYAN}kubectl get pods -n hipster-frontend${RESET}"
echo -e "  ${CYAN}kubectl get pods -n hipster-database${RESET}"
