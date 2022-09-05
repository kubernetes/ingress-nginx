package dataplane

import (
	"bytes"
	"html/template"
	"os"

	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/pkg/util/file"
)

const zipkinTmpl = `{
	"service_name": "{{ .ZipkinServiceName }}",
	"collector_host": "{{ .ZipkinCollectorHost }}",
	"collector_port": {{ .ZipkinCollectorPort }},
	"sample_rate": {{ .ZipkinSampleRate }}
  }`

const jaegerTmpl = `{
	"service_name": "{{ .JaegerServiceName }}",
	"propagation_format": "{{ .JaegerPropagationFormat }}",
	"sampler": {
	  "type": "{{ .JaegerSamplerType }}",
	  "param": {{ .JaegerSamplerParam }},
	  "samplingServerURL": "{{ .JaegerSamplerHost }}:{{ .JaegerSamplerPort }}/sampling"
	},
	"reporter": {
	  "endpoint": "{{ .JaegerEndpoint }}",
	  "localAgentHostPort": "{{ .JaegerCollectorHost }}:{{ .JaegerCollectorPort }}"
	},
	"headers": {
	  "TraceContextHeaderName": "{{ .JaegerTraceContextHeaderName }}",
	  "jaegerDebugHeader": "{{ .JaegerDebugHeader }}",
	  "jaegerBaggageHeader": "{{ .JaegerBaggageHeader }}",
	  "traceBaggageHeaderPrefix": "{{ .JaegerTraceBaggageHeaderPrefix }}"
	}
  }`

const datadogTmpl = `{
	"service": "{{ .DatadogServiceName }}",
	"agent_host": "{{ .DatadogCollectorHost }}",
	"agent_port": {{ .DatadogCollectorPort }},
	"environment": "{{ .DatadogEnvironment }}",
	"operation_name_override": "{{ .DatadogOperationNameOverride }}",
	"sample_rate": {{ .DatadogSampleRate }},
	"dd.priority.sampling": {{ .DatadogPrioritySampling }}
  }`

func createOpentracingCfg(cfg config.Configuration) error {
	var tmpl *template.Template
	var err error

	if cfg.ZipkinCollectorHost != "" {
		tmpl, err = template.New("zipkin").Parse(zipkinTmpl)
		if err != nil {
			return err
		}
	} else if cfg.JaegerCollectorHost != "" || cfg.JaegerEndpoint != "" {
		tmpl, err = template.New("jaeger").Parse(jaegerTmpl)
		if err != nil {
			return err
		}
	} else if cfg.DatadogCollectorHost != "" {
		tmpl, err = template.New("datadog").Parse(datadogTmpl)
		if err != nil {
			return err
		}
	} else {
		tmpl, _ = template.New("empty").Parse("{}")
	}

	tmplBuf := bytes.NewBuffer(make([]byte, 0))
	err = tmpl.Execute(tmplBuf, cfg)
	if err != nil {
		return err
	}

	// Expand possible environment variables before writing the configuration to file.
	expanded := os.ExpandEnv(tmplBuf.String())

	return os.WriteFile("/etc/nginx/opentracing.json", []byte(expanded), file.ReadWriteByUser)
}
