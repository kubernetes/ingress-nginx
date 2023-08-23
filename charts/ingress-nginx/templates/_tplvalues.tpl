{{/* vim: set filetype=mustache: */}}
{{/*
Renders a value that contains template perhaps with scope if the scope is present.
Usage:
{{ include "tplvalues.render" ( dict "value" .Values.path.to.the.Value "context" $ ) }}
{{ include "tplvalues.render" ( dict "value" .Values.path.to.the.Value "context" $ "scope" $app ) }}
*/}}
{{- define "tplvalues.render" -}}
{{- if .scope }}
  {{- if typeIs "string" .value }}
    {{- tpl (cat "{{- with $.RelativeScope -}}" .value  "{{- end }}") (merge (dict "RelativeScope" .scope) .context) }}
  {{- else }}
    {{- tpl (cat "{{- with $.RelativeScope -}}" (.value | toYaml)  "{{- end }}") (merge (dict "RelativeScope" .scope) .context) }}
  {{- end }}
{{- else }}
  {{- if typeIs "string" .value }}
    {{- tpl .value .context }}
  {{- else }}
    {{- tpl (.value | toYaml) .context }}
  {{- end }}
{{- end -}}
{{- end -}}