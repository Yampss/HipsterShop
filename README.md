# HipsterShop

Minimal microservicesapp for Kubernetes.

## What is included

- Polyglot services under src
- Kubernetes manifests under kubernetes-manifests
- Optional observability install script

## Quick start

Prerequisites:
- Docker
- Kubernetes cluster
- kubectl
- Helm (optional, for observability)

Build and push images (update image names in dockerpush.sh first):

./dockerpush.sh

Deploy core services:

kubectl apply -k kubernetes-manifests

Deploy optional components:


kubectl apply -f kubernetes-manifests/postgres.yaml

Check workload status:

kubectl get pods
kubectl get svc

## Observability (optional)

Install Prometheus, Grafana, and Jaeger:

./install_observability.sh

Then follow:
- observability_walkthrough.md

## Docs

Service documentation is in:
- documentation/
