{{- define "ingress-nginx.params" -}}
- /nginx-ingress-controller
{{- if .Values.controller.enableAnnotationValidations }}
- --enable-annotation-validation=true
{{- end }}
{{- if .Values.defaultBackend.enabled }}
- --default-backend-service=$(POD_NAMESPACE)/{{ include "ingress-nginx.defaultBackend.fullname" . }}
{{- end }}
{{- if and .Values.controller.publishService.enabled .Values.controller.service.enabled }}
{{- if .Values.controller.service.external.enabled }}
- --publish-service={{ template "ingress-nginx.controller.publishServicePath" . }}
{{- else if .Values.controller.service.internal.enabled }}
- --publish-service={{ template "ingress-nginx.controller.publishServicePath" . }}-internal
{{- end }}
{{- end }}
- --election-id={{ include "ingress-nginx.controller.electionID" . }}
- --controller-class={{ .Values.controller.ingressClassResource.controllerValue }}
{{- if .Values.controller.ingressClass }}
- --ingress-class={{ .Values.controller.ingressClass }}
{{- end }}
- --configmap={{ default "$(POD_NAMESPACE)" .Values.controller.configMapNamespace }}/{{ include "ingress-nginx.controller.fullname" . }}
{{- if .Values.tcp }}
- --tcp-services-configmap={{ default "$(POD_NAMESPACE)" .Values.controller.tcp.configMapNamespace }}/{{ include "ingress-nginx.fullname" . }}-tcp
{{- end }}
{{- if .Values.udp }}
- --udp-services-configmap={{ default "$(POD_NAMESPACE)" .Values.controller.udp.configMapNamespace }}/{{ include "ingress-nginx.fullname" . }}-udp
{{- end }}
{{- if .Values.controller.scope.enabled }}
- --watch-namespace={{ default "$(POD_NAMESPACE)" .Values.controller.scope.namespace }}
{{- end }}
{{- if and (not .Values.controller.scope.enabled) .Values.controller.scope.namespaceSelector }}
- --watch-namespace-selector={{ .Values.controller.scope.namespaceSelector }}
{{- end }}
{{- if and .Values.controller.reportNodeInternalIp .Values.controller.hostNetwork }}
- --report-node-internal-ip-address={{ .Values.controller.reportNodeInternalIp }}
{{- end }}
{{- if .Values.controller.admissionWebhooks.enabled }}
- --validating-webhook=:{{ .Values.controller.admissionWebhooks.port }}
- --validating-webhook-certificate={{ .Values.controller.admissionWebhooks.certificate }}
- --validating-webhook-key={{ .Values.controller.admissionWebhooks.key }}
{{- end }}
{{- if .Values.controller.maxmindLicenseKey }}
- --maxmind-license-key={{ .Values.controller.maxmindLicenseKey }}
{{- end }}
{{- if .Values.controller.healthCheckHost }}
- --healthz-host={{ .Values.controller.healthCheckHost }}
{{- end }}
{{- if not (eq .Values.controller.healthCheckPath "/healthz") }}
- --health-check-path={{ .Values.controller.healthCheckPath }}
{{- end }}
{{- if .Values.controller.ingressClassByName }}
- --ingress-class-by-name=true
{{- end }}
{{- if .Values.controller.watchIngressWithoutClass }}
- --watch-ingress-without-class=true
{{- end }}
{{- if not .Values.controller.metrics.enabled }}
- --enable-metrics={{ .Values.controller.metrics.enabled }}
{{- end }}
{{- if .Values.controller.enableTopologyAwareRouting }}
- --enable-topology-aware-routing=true
{{- end }}
{{- if .Values.controller.disableLeaderElection }}
- --disable-leader-election=true
{{- end }}
{{- if .Values.controller.electionTTL }}
- --election-ttl={{ .Values.controller.electionTTL }}
{{- end }}
{{- range $key, $value := .Values.controller.extraArgs }}
{{- /* Accept keys without values or with false as value */}}
{{- if eq ($value | quote | len) 2 }}
- --{{ $key }}
{{- else }}
- --{{ $key }}={{ $value }}
{{- end }}
{{- end }}
{{- end -}}
