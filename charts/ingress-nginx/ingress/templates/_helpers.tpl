{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "ingress.name" -}}
{{- if .Values.nameSuffix -}}
{{- printf "%s-%s" .Release.Name .Values.nameSuffix | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- default .Release.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "ingress.service.name" -}}
{{- if .Values.service.nameSuffix -}}
{{- printf "%s-%s" .Release.Name .Values.service.nameSuffix | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- default .Release.Name .Values.service.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ingress.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "ingress.labels" -}}
app.kubernetes.io/name: {{ include "ingress.name" . }}
helm.sh/chart: {{ include "ingress.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}
