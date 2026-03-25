#!/bin/bash
# install_observability.sh
# This script installs Prometheus and Jaeger OpenTelemetry Collector via Helm into your Kubernetes cluster.

echo "========================================="
echo " Installing Jaeger OpenTelemetry Collector "
echo "========================================="
# Add Jaeger Helm Repository
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
helm repo update

# Install Jaeger into the 'observability' namespace
# We use custom values to ensure the OpenTelemetry Collector receiver is enabled out of the box
helm install jaeger jaegertracing/jaeger \
  --namespace observability \
  --create-namespace \
  --set provisionDataStore.cassandra=false \
  --set allInOne.enabled=true \
  --set storage.type=memory \
  --set agent.enabled=false \
  --set collector.enabled=false \
  --set query.enabled=false

echo ""
echo "========================================="
echo " Installing Prometheus / Grafana Stack     "
echo "========================================="
# Add Prometheus Helm Repository
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install kube-prometheus-stack
helm install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace
  
echo ""
echo "========================================="
echo " Observability Stack Installation Complete!"
echo "========================================="
echo "Please wait a few moments for the pods to initialize."
echo ""
echo "To view Grafana dashboards:"
echo "  kubectl port-forward svc/monitoring-grafana 8080:80 -n monitoring"
echo "  (Login credentials normally: admin / prom-operator)"
echo ""
echo "To view Jaeger Distributed Traces:"
echo "  kubectl port-forward svc/jaeger 16686:16686 -n observability"
