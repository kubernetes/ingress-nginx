{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "ingress-nginx.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ingress-nginx.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "ingress-nginx.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Expand the namespace of the release.
Allows overriding it for multi-namespace deployments in combined charts.
*/}}
{{- define "ingress-nginx.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Controller container security context.
*/}}
{{- define "ingress-nginx.controller.containerSecurityContext" -}}
{{- if .Values.controller.containerSecurityContext -}}
{{- toYaml .Values.controller.containerSecurityContext -}}
{{- else -}}
runAsNonRoot: {{ .Values.controller.image.runAsNonRoot }}
runAsUser: {{ .Values.controller.image.runAsUser }}
allowPrivilegeEscalation: {{ or .Values.controller.image.allowPrivilegeEscalation .Values.controller.image.chroot }}
{{- if .Values.controller.image.seccompProfile }}
seccompProfile: {{ toYaml .Values.controller.image.seccompProfile | nindent 2 }}
{{- end }}
capabilities:
  drop:
  - ALL
  add:
  - NET_BIND_SERVICE
  {{- if .Values.controller.image.chroot }}
  {{- if .Values.controller.image.seccompProfile }}
  - SYS_ADMIN
  {{- end }}
  - SYS_CHROOT
  {{- end }}
readOnlyRootFilesystem: {{ .Values.controller.image.readOnlyRootFilesystem }}
{{- end -}}
{{- end -}}

{{/*
Get specific image
*/}}
{{- define "ingress-nginx.image" -}}
{{- if .chroot -}}
{{- printf "%s-chroot" .image -}}
{{- else -}}
{{- printf "%s" .image -}}
{{- end }}
{{- end -}}

{{/*
Get specific image digest
*/}}
{{- define "ingress-nginx.imageDigest" -}}
{{- if .chroot -}}
{{- if .digestChroot -}}
{{- printf "@%s" .digestChroot -}}
{{- end }}
{{- else -}}
{{ if .digest -}}
{{- printf "@%s" .digest -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create a default fully qualified controller name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "ingress-nginx.controller.fullname" -}}
{{- printf "%s-%s" (include "ingress-nginx.fullname" .) .Values.controller.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Construct a unique electionID.
Users can provide an override for an explicit electionID if they want via `.Values.controller.electionID`
*/}}
{{- define "ingress-nginx.controller.electionID" -}}
{{- $defElectionID := printf "%s-leader" (include "ingress-nginx.fullname" .) -}}
{{- $electionID := default $defElectionID .Values.controller.electionID -}}
{{- print $electionID -}}
{{- end -}}

{{/*
Construct the path for the publish-service.

By convention this will simply use the <namespace>/<controller-name> to match the name of the
service generated.

Users can provide an override for an explicit service they want bound via `.Values.controller.publishService.pathOverride`
*/}}
{{- define "ingress-nginx.controller.publishServicePath" -}}
{{- $defServiceName := printf "%s/%s" "$(POD_NAMESPACE)" (include "ingress-nginx.controller.fullname" .) -}}
{{- $servicePath := default $defServiceName .Values.controller.publishService.pathOverride }}
{{- print $servicePath | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "ingress-nginx.labels" -}}
helm.sh/chart: {{ include "ingress-nginx.chart" . }}
{{ include "ingress-nginx.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/part-of: {{ template "ingress-nginx.name" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- if .Values.commonLabels}}
{{ toYaml .Values.commonLabels }}
{{- end }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "ingress-nginx.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ingress-nginx.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the controller service account to use
*/}}
{{- define "ingress-nginx.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "ingress-nginx.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Create a default fully qualified admission webhook name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "ingress-nginx.admissionWebhooks.fullname" -}}
{{- printf "%s-%s" (include "ingress-nginx.fullname" .) .Values.controller.admissionWebhooks.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the admission webhook patch job service account to use
*/}}
{{- define "ingress-nginx.admissionWebhooks.patch.serviceAccountName" -}}
{{- if .Values.controller.admissionWebhooks.patch.serviceAccount.create -}}
    {{ default (include "ingress-nginx.admissionWebhooks.fullname" .) .Values.controller.admissionWebhooks.patch.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.controller.admissionWebhooks.patch.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Create a default fully qualified admission webhook secret creation job name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "ingress-nginx.admissionWebhooks.createSecretJob.fullname" -}}
{{- printf "%s-%s" (include "ingress-nginx.admissionWebhooks.fullname" .) .Values.controller.admissionWebhooks.createSecretJob.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified admission webhook patch job name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "ingress-nginx.admissionWebhooks.patchWebhookJob.fullname" -}}
{{- printf "%s-%s" (include "ingress-nginx.admissionWebhooks.fullname" .) .Values.controller.admissionWebhooks.patchWebhookJob.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified default backend name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "ingress-nginx.defaultBackend.fullname" -}}
{{- printf "%s-%s" (include "ingress-nginx.fullname" .) .Values.defaultBackend.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the backend service account to use - only used when podsecuritypolicy is also enabled
*/}}
{{- define "ingress-nginx.defaultBackend.serviceAccountName" -}}
{{- if .Values.defaultBackend.serviceAccount.create -}}
    {{ default (printf "%s-backend" (include "ingress-nginx.fullname" .)) .Values.defaultBackend.serviceAccount.name }}
{{- else -}}
    {{ default "default-backend" .Values.defaultBackend.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Default backend container security context.
*/}}
{{- define "ingress-nginx.defaultBackend.containerSecurityContext" -}}
{{- if .Values.defaultBackend.containerSecurityContext -}}
{{- toYaml .Values.defaultBackend.containerSecurityContext -}}
{{- else -}}
runAsNonRoot: {{ .Values.defaultBackend.image.runAsNonRoot }}
runAsUser: {{ .Values.defaultBackend.image.runAsUser }}
allowPrivilegeEscalation: {{ .Values.defaultBackend.image.allowPrivilegeEscalation }}
{{- if .Values.defaultBackend.image.seccompProfile }}
seccompProfile: {{ toYaml .Values.defaultBackend.image.seccompProfile | nindent 2 }}
{{- end }}
capabilities:
  drop:
  - ALL
readOnlyRootFilesystem: {{ .Values.defaultBackend.image.readOnlyRootFilesystem }}
{{- end -}}
{{- end -}}

{{/*
Return the appropriate apiGroup for PodSecurityPolicy.
*/}}
{{- define "podSecurityPolicy.apiGroup" -}}
{{- if semverCompare ">=1.14-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "policy" -}}
{{- else -}}
{{- print "extensions" -}}
{{- end -}}
{{- end -}}

{{/*
Extra modules.
*/}}
{{- define "extraModules" -}}
- name: {{ .name }}
  {{- with .image }}
  image: {{ if .repository }}{{ .repository }}{{ else }}{{ .registry }}/{{ .image }}{{ end }}:{{ .tag }}{{ if .digest }}@{{ .digest }}{{ end }}
  command:
  {{- if .distroless }}
    - /init_module
  {{- else }}
    - sh
    - -c
    - /usr/local/bin/init_module.sh
  {{- end }}
  {{- end }}
  {{- if .containerSecurityContext }}
  securityContext: {{ toYaml .containerSecurityContext | nindent 4 }}
  {{- end }}
  {{- if .resources }}
  resources: {{ toYaml .resources | nindent 4 }}
  {{- end }}
  volumeMounts:
    - name: modules
      mountPath: /modules_mount
{{- end -}}
