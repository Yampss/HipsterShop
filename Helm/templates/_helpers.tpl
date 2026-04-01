{{/*
===========================================================
  HipsterShop Helm Helpers
  Reusable template snippets for all chart templates.
===========================================================
*/}}

{{/*
Chart name (truncated to 63 chars).
*/}}
{{- define "hipstershop.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Standard labels applied to every resource.
*/}}
{{- define "hipstershop.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}

{{/*
Release namespace — used throughout templates.
Usage: {{ template "hipstershop.namespace" . }}
*/}}
{{- define "hipstershop.namespace" -}}
{{- .Release.Namespace }}
{{- end }}

{{/*
MongoDB headless service hostname list for replica-set connection strings.
Returns the comma-separated host list for use inside MONGO_URI values.
Usage: {{ include "hipstershop.mongoHosts" . }}
*/}}
{{- define "hipstershop.mongoHosts" -}}
mongodb-0.mongo-headless.{{ .Release.Namespace }}.svc.cluster.local:27017,mongodb-1.mongo-headless.{{ .Release.Namespace }}.svc.cluster.local:27017,mongodb-2.mongo-headless.{{ .Release.Namespace }}.svc.cluster.local:27017
{{- end }}

{{/*
MongoDB full URI builder.
Args (via dict): .db = database name, .authSource = auth source db.
Usage:
  value: "mongodb://$(MONGO_USERNAME):$(MONGO_PASSWORD)@{{ include "hipstershop.mongoHosts" . }}/{{ .db }}?replicaSet=rs0&authSource={{ .authSource }}&authMechanism=SCRAM-SHA-256"
Note: $(MONGO_USERNAME) & $(MONGO_PASSWORD) are K8s env-var references
      resolved at container start — NOT Helm template expressions.
*/}}

{{/*
RollingUpdate strategy block — shared across all Deployments.
*/}}
{{- define "hipstershop.rollingUpdate" -}}
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxUnavailable: {{ .Values.rollingUpdate.maxUnavailable }}
    maxSurge: {{ .Values.rollingUpdate.maxSurge }}
{{- end }}

{{/*
Standard resource block for most services.
Usage: {{ include "hipstershop.resources" .Values.resources.standard }}
*/}}
{{- define "hipstershop.resourcesStandard" -}}
resources:
  requests:
    cpu: 100m
    memory: 64Mi
  limits:
    cpu: 200m
    memory: 128Mi
{{- end }}
