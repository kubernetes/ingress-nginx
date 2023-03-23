//go:build mage

/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	semver "github.com/blang/semver/v4"
	"github.com/helm/helm/pkg/chartutil"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	yamlpath "github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

const HelmChartPath = "charts/ingress-nginx/Chart.yaml"
const HelmChartValues = "charts/ingress-nginx/values.yaml"

type Helm mg.Namespace

// UpdateAppVersion Updates the Helm App Version of Ingress Nginx Controller
func (Helm) UpdateAppVersion() {
	updateAppVersion()
}

func updateAppVersion() {

}

// UpdateVersion Update Helm Version of the Chart
func (Helm) UpdateVersion(version string) {
	updateVersion(version)
}

func currentChartVersion() string {
	chart, err := chartutil.LoadChartfile(HelmChartPath)
	CheckIfError(err, "HELM Could not Load Chart")
	return chart.Version
}

func currentChartAppVersion() string {
	chart, err := chartutil.LoadChartfile(HelmChartPath)
	CheckIfError(err, "HELM Could not Load Chart")
	return chart.AppVersion
}

func updateVersion(version string) {
	Info("HELM Reading File %v", HelmChartPath)

	chart, err := chartutil.LoadChartfile(HelmChartPath)
	CheckIfError(err, "HELM Could not Load Chart")

	//Get the current tag
	//appVersionV, err := getIngressNGINXVersion()
	//CheckIfError(err, "HELM Issue Retrieving the Current Ingress Nginx Version")

	//remove the v from TAG
	appVersion := version

	Info("HELM Ingress-Nginx App Version: %s Chart AppVersion: %s", appVersion, chart.AppVersion)
	if appVersion == chart.AppVersion {
		Warning("HELM Ingress NGINX Version didnt change Ingress-Nginx App Version: %s Chart AppVersion: %s", appVersion, chart.AppVersion)
		return
	}

	//Update the helm chart
	chart.AppVersion = appVersion
	cTag, err := semver.Make(chart.Version)
	CheckIfError(err, "HELM Creating Chart Version: %v", err)

	if err = cTag.IncrementPatch(); err != nil {
		ErrorF("HELM Incrementing Chart Version: %v", err)
		os.Exit(1)
	}
	chart.Version = cTag.String()
	Debug("HELM Updated Chart Version: %v", chart.Version)

	err = chartutil.SaveChartfile(HelmChartPath, chart)
	CheckIfError(err, "HELM Saving new Chart")
}

func updateChartReleaseNotes(releasesNotes []string) {
	Info("HELM Updating the Chart Release notes")
	chart, err := chartutil.LoadChartfile(HelmChartPath)
	CheckIfError(err, "HELM Could not Load Chart to update release notes %s", HelmChartPath)
	var releaseNoteString string
	for i := range releasesNotes {
		releaseNoteString = fmt.Sprintf("%s - %s\n", releaseNoteString, releasesNotes[i])
	}
	Info("HELM Release note string %s", releaseNoteString)
	chart.Annotations["artifacthub.io/changes"] = releaseNoteString
	err = chartutil.SaveChartfile(HelmChartPath, chart)
	CheckIfError(err, "HELM Saving updated release notes for Chart")
}

func UpdateChartChangelog() {

}

// UpdateChartValue Updates the Helm ChartValue
func (Helm) UpdateChartValue(key, value string) {
	updateChartValue(key, value)
}

func updateChartValue(key, value string) {
	Info("HELM Updating Chart %s %s:%s", HelmChartValues, key, value)

	//read current values.yaml
	data, err := os.ReadFile(HelmChartValues)
	CheckIfError(err, "HELM Could not Load Helm Chart Values files %s", HelmChartValues)

	//var valuesStruct IngressChartValue
	var n yaml.Node
	CheckIfError(yaml.Unmarshal(data, &n), "HELM Could not Unmarshal %s", HelmChartValues)

	//update value
	//keyParse := parsePath(key)
	p, err := yamlpath.NewPath(key)
	CheckIfError(err, "HELM cannot create path")

	q, err := p.Find(&n)
	CheckIfError(err, "HELM unexpected error finding path")

	for _, i := range q {
		Info("HELM Found %s at %s", i.Value, key)
		i.Value = value
		Info("HELM Updated %s at %s", i.Value, key)
	}

	//// write to file
	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(&n)
	CheckIfError(err, "HELM Could not Marshal new Values file")
	err = os.WriteFile(HelmChartValues, b.Bytes(), 0644)
	CheckIfError(err, "HELM Could not write new Values file to %s", HelmChartValues)

	Info("HELM Ingress Nginx Helm Chart update %s %s", key, value)
}

func (Helm) Helmdocs() error {
	return runHelmDocs()
}
func runHelmDocs() error {
	err := installHelmDocs()
	if err != nil {
		return err
	}
	err = sh.RunV("helm-docs", "--chart-search-root=${PWD}/charts")
	if err != nil {
		return err
	}
	return nil
}

func installHelmDocs() error {
	Info("HELM Install HelmDocs")
	var g0 = sh.RunCmd("go")

	err := g0("install", "github.com/norwoodj/helm-docs/cmd/helm-docs@v1.11.0")
	if err != nil {
		return err
	}
	return nil
}
func parsePath(key string) []string { return strings.Split(key, ".") }

func updateHelmDocs() {

}

type IngressChartValue struct {
	CommonLabels struct {
	} `yaml:"commonLabels"`
	Controller struct {
		Name  string `yaml:"name"`
		Image struct {
			Chroot                   bool   `yaml:"chroot"`
			Registry                 string `yaml:"registry"`
			Image                    string `yaml:"image"`
			Tag                      string `yaml:"tag"`
			Digest                   string `yaml:"digest"`
			DigestChroot             string `yaml:"digestChroot"`
			PullPolicy               string `yaml:"pullPolicy"`
			RunAsUser                int    `yaml:"runAsUser"`
			AllowPrivilegeEscalation bool   `yaml:"allowPrivilegeEscalation"`
		} `yaml:"image"`
		ExistingPsp   string `yaml:"existingPsp"`
		ContainerName string `yaml:"containerName"`
		ContainerPort struct {
			HTTP  int `yaml:"http"`
			HTTPS int `yaml:"https"`
		} `yaml:"containerPort"`
		Config struct {
		} `yaml:"config"`
		ConfigAnnotations struct {
		} `yaml:"configAnnotations"`
		ProxySetHeaders struct {
		} `yaml:"proxySetHeaders"`
		AddHeaders struct {
		} `yaml:"addHeaders"`
		DNSConfig struct {
		} `yaml:"dnsConfig"`
		Hostname struct {
		} `yaml:"hostname"`
		DNSPolicy                string `yaml:"dnsPolicy"`
		ReportNodeInternalIP     bool   `yaml:"reportNodeInternalIp"`
		WatchIngressWithoutClass bool   `yaml:"watchIngressWithoutClass"`
		IngressClassByName       bool   `yaml:"ingressClassByName"`
		AllowSnippetAnnotations  bool   `yaml:"allowSnippetAnnotations"`
		HostNetwork              bool   `yaml:"hostNetwork"`
		HostPort                 struct {
			Enabled bool `yaml:"enabled"`
			Ports   struct {
				HTTP  int `yaml:"http"`
				HTTPS int `yaml:"https"`
			} `yaml:"ports"`
		} `yaml:"hostPort"`
		ElectionID           string `yaml:"electionID"`
		IngressClassResource struct {
			Name            string `yaml:"name"`
			Enabled         bool   `yaml:"enabled"`
			Default         bool   `yaml:"default"`
			ControllerValue string `yaml:"controllerValue"`
			Parameters      struct {
			} `yaml:"parameters"`
		} `yaml:"ingressClassResource"`
		IngressClass string `yaml:"ingressClass"`
		PodLabels    struct {
		} `yaml:"podLabels"`
		PodSecurityContext struct {
		} `yaml:"podSecurityContext"`
		Sysctls struct {
		} `yaml:"sysctls"`
		PublishService struct {
			Enabled      bool   `yaml:"enabled"`
			PathOverride string `yaml:"pathOverride"`
		} `yaml:"publishService"`
		Scope struct {
			Enabled           bool   `yaml:"enabled"`
			Namespace         string `yaml:"namespace"`
			NamespaceSelector string `yaml:"namespaceSelector"`
		} `yaml:"scope"`
		ConfigMapNamespace string `yaml:"configMapNamespace"`
		TCP                struct {
			ConfigMapNamespace string `yaml:"configMapNamespace"`
			Annotations        struct {
			} `yaml:"annotations"`
		} `yaml:"tcp"`
		UDP struct {
			ConfigMapNamespace string `yaml:"configMapNamespace"`
			Annotations        struct {
			} `yaml:"annotations"`
		} `yaml:"udp"`
		MaxmindLicenseKey string `yaml:"maxmindLicenseKey"`
		ExtraArgs         struct {
		} `yaml:"extraArgs"`
		ExtraEnvs   []interface{} `yaml:"extraEnvs"`
		Kind        string        `yaml:"kind"`
		Annotations struct {
		} `yaml:"annotations"`
		Labels struct {
		} `yaml:"labels"`
		UpdateStrategy struct {
		} `yaml:"updateStrategy"`
		MinReadySeconds int           `yaml:"minReadySeconds"`
		Tolerations     []interface{} `yaml:"tolerations"`
		Affinity        struct {
		} `yaml:"affinity"`
		TopologySpreadConstraints     []interface{} `yaml:"topologySpreadConstraints"`
		TerminationGracePeriodSeconds int           `yaml:"terminationGracePeriodSeconds"`
		NodeSelector                  struct {
			KubernetesIoOs string `yaml:"kubernetes.io/os"`
		} `yaml:"nodeSelector"`
		LivenessProbe struct {
			HTTPGet struct {
				Path   string `yaml:"path"`
				Port   int    `yaml:"port"`
				Scheme string `yaml:"scheme"`
			} `yaml:"httpGet"`
			InitialDelaySeconds int `yaml:"initialDelaySeconds"`
			PeriodSeconds       int `yaml:"periodSeconds"`
			TimeoutSeconds      int `yaml:"timeoutSeconds"`
			SuccessThreshold    int `yaml:"successThreshold"`
			FailureThreshold    int `yaml:"failureThreshold"`
		} `yaml:"livenessProbe"`
		ReadinessProbe struct {
			HTTPGet struct {
				Path   string `yaml:"path"`
				Port   int    `yaml:"port"`
				Scheme string `yaml:"scheme"`
			} `yaml:"httpGet"`
			InitialDelaySeconds int `yaml:"initialDelaySeconds"`
			PeriodSeconds       int `yaml:"periodSeconds"`
			TimeoutSeconds      int `yaml:"timeoutSeconds"`
			SuccessThreshold    int `yaml:"successThreshold"`
			FailureThreshold    int `yaml:"failureThreshold"`
		} `yaml:"readinessProbe"`
		HealthCheckPath string `yaml:"healthCheckPath"`
		HealthCheckHost string `yaml:"healthCheckHost"`
		PodAnnotations  struct {
		} `yaml:"podAnnotations"`
		ReplicaCount int `yaml:"replicaCount"`
		MinAvailable int `yaml:"minAvailable"`
		Resources    struct {
			Requests struct {
				CPU    string `yaml:"cpu"`
				Memory string `yaml:"memory"`
			} `yaml:"requests"`
		} `yaml:"resources"`
		Autoscaling struct {
			APIVersion  string `yaml:"apiVersion"`
			Enabled     bool   `yaml:"enabled"`
			Annotations struct {
			} `yaml:"annotations"`
			MinReplicas                       int `yaml:"minReplicas"`
			MaxReplicas                       int `yaml:"maxReplicas"`
			TargetCPUUtilizationPercentage    int `yaml:"targetCPUUtilizationPercentage"`
			TargetMemoryUtilizationPercentage int `yaml:"targetMemoryUtilizationPercentage"`
			Behavior                          struct {
			} `yaml:"behavior"`
		} `yaml:"autoscaling"`
		AutoscalingTemplate []interface{} `yaml:"autoscalingTemplate"`
		Keda                struct {
			APIVersion                    string `yaml:"apiVersion"`
			Enabled                       bool   `yaml:"enabled"`
			MinReplicas                   int    `yaml:"minReplicas"`
			MaxReplicas                   int    `yaml:"maxReplicas"`
			PollingInterval               int    `yaml:"pollingInterval"`
			CooldownPeriod                int    `yaml:"cooldownPeriod"`
			RestoreToOriginalReplicaCount bool   `yaml:"restoreToOriginalReplicaCount"`
			ScaledObject                  struct {
				Annotations struct {
				} `yaml:"annotations"`
			} `yaml:"scaledObject"`
			Triggers []interface{} `yaml:"triggers"`
			Behavior struct {
			} `yaml:"behavior"`
		} `yaml:"keda"`
		EnableMimalloc bool `yaml:"enableMimalloc"`
		CustomTemplate struct {
			ConfigMapName string `yaml:"configMapName"`
			ConfigMapKey  string `yaml:"configMapKey"`
		} `yaml:"customTemplate"`
		Service struct {
			Enabled     bool `yaml:"enabled"`
			AppProtocol bool `yaml:"appProtocol"`
			Annotations struct {
			} `yaml:"annotations"`
			Labels struct {
			} `yaml:"labels"`
			ExternalIPs              []interface{} `yaml:"externalIPs"`
			LoadBalancerIP           string        `yaml:"loadBalancerIP"`
			LoadBalancerSourceRanges []interface{} `yaml:"loadBalancerSourceRanges"`
			EnableHTTP               bool          `yaml:"enableHttp"`
			EnableHTTPS              bool          `yaml:"enableHttps"`
			IPFamilyPolicy           string        `yaml:"ipFamilyPolicy"`
			IPFamilies               []string      `yaml:"ipFamilies"`
			Ports                    struct {
				HTTP  int `yaml:"http"`
				HTTPS int `yaml:"https"`
			} `yaml:"ports"`
			TargetPorts struct {
				HTTP  string `yaml:"http"`
				HTTPS string `yaml:"https"`
			} `yaml:"targetPorts"`
			Type      string `yaml:"type"`
			NodePorts struct {
				HTTP  string `yaml:"http"`
				HTTPS string `yaml:"https"`
				TCP   struct {
				} `yaml:"tcp"`
				UDP struct {
				} `yaml:"udp"`
			} `yaml:"nodePorts"`
			External struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"external"`
			Internal struct {
				Enabled     bool `yaml:"enabled"`
				Annotations struct {
				} `yaml:"annotations"`
				LoadBalancerSourceRanges []interface{} `yaml:"loadBalancerSourceRanges"`
			} `yaml:"internal"`
		} `yaml:"service"`
		ShareProcessNamespace bool          `yaml:"shareProcessNamespace"`
		ExtraContainers       []interface{} `yaml:"extraContainers"`
		ExtraVolumeMounts     []interface{} `yaml:"extraVolumeMounts"`
		ExtraVolumes          []interface{} `yaml:"extraVolumes"`
		ExtraInitContainers   []interface{} `yaml:"extraInitContainers"`
		ExtraModules          []interface{} `yaml:"extraModules"`
		Opentelemetry         struct {
			Enabled                  bool   `yaml:"enabled"`
			Image                    string `yaml:"image"`
			ContainerSecurityContext struct {
				AllowPrivilegeEscalation bool `yaml:"allowPrivilegeEscalation"`
			} `yaml:"containerSecurityContext"`
		} `yaml:"opentelemetry"`
		AdmissionWebhooks struct {
			Annotations struct {
			} `yaml:"annotations"`
			Enabled           bool          `yaml:"enabled"`
			ExtraEnvs         []interface{} `yaml:"extraEnvs"`
			FailurePolicy     string        `yaml:"failurePolicy"`
			Port              int           `yaml:"port"`
			Certificate       string        `yaml:"certificate"`
			Key               string        `yaml:"key"`
			NamespaceSelector struct {
			} `yaml:"namespaceSelector"`
			ObjectSelector struct {
			} `yaml:"objectSelector"`
			Labels struct {
			} `yaml:"labels"`
			ExistingPsp          string `yaml:"existingPsp"`
			NetworkPolicyEnabled bool   `yaml:"networkPolicyEnabled"`
			Service              struct {
				Annotations struct {
				} `yaml:"annotations"`
				ExternalIPs              []interface{} `yaml:"externalIPs"`
				LoadBalancerSourceRanges []interface{} `yaml:"loadBalancerSourceRanges"`
				ServicePort              int           `yaml:"servicePort"`
				Type                     string        `yaml:"type"`
			} `yaml:"service"`
			CreateSecretJob struct {
				SecurityContext struct {
					AllowPrivilegeEscalation bool `yaml:"allowPrivilegeEscalation"`
				} `yaml:"securityContext"`
				Resources struct {
				} `yaml:"resources"`
			} `yaml:"createSecretJob"`
			PatchWebhookJob struct {
				SecurityContext struct {
					AllowPrivilegeEscalation bool `yaml:"allowPrivilegeEscalation"`
				} `yaml:"securityContext"`
				Resources struct {
				} `yaml:"resources"`
			} `yaml:"patchWebhookJob"`
			Patch struct {
				Enabled bool `yaml:"enabled"`
				Image   struct {
					Registry   string `yaml:"registry"`
					Image      string `yaml:"image"`
					Tag        string `yaml:"tag"`
					Digest     string `yaml:"digest"`
					PullPolicy string `yaml:"pullPolicy"`
				} `yaml:"image"`
				PriorityClassName string `yaml:"priorityClassName"`
				PodAnnotations    struct {
				} `yaml:"podAnnotations"`
				NodeSelector struct {
					KubernetesIoOs string `yaml:"kubernetes.io/os"`
				} `yaml:"nodeSelector"`
				Tolerations []interface{} `yaml:"tolerations"`
				Labels      struct {
				} `yaml:"labels"`
				SecurityContext struct {
					RunAsNonRoot bool `yaml:"runAsNonRoot"`
					RunAsUser    int  `yaml:"runAsUser"`
					FsGroup      int  `yaml:"fsGroup"`
				} `yaml:"securityContext"`
			} `yaml:"patch"`
			CertManager struct {
				Enabled  bool `yaml:"enabled"`
				RootCert struct {
					Duration string `yaml:"duration"`
				} `yaml:"rootCert"`
				AdmissionCert struct {
					Duration string `yaml:"duration"`
				} `yaml:"admissionCert"`
			} `yaml:"certManager"`
		} `yaml:"admissionWebhooks"`
		Metrics struct {
			Port     int    `yaml:"port"`
			PortName string `yaml:"portName"`
			Enabled  bool   `yaml:"enabled"`
			Service  struct {
				Annotations struct {
				} `yaml:"annotations"`
				ExternalIPs              []interface{} `yaml:"externalIPs"`
				LoadBalancerSourceRanges []interface{} `yaml:"loadBalancerSourceRanges"`
				ServicePort              int           `yaml:"servicePort"`
				Type                     string        `yaml:"type"`
			} `yaml:"service"`
			ServiceMonitor struct {
				Enabled          bool `yaml:"enabled"`
				AdditionalLabels struct {
				} `yaml:"additionalLabels"`
				Namespace         string `yaml:"namespace"`
				NamespaceSelector struct {
				} `yaml:"namespaceSelector"`
				ScrapeInterval    string        `yaml:"scrapeInterval"`
				TargetLabels      []interface{} `yaml:"targetLabels"`
				Relabelings       []interface{} `yaml:"relabelings"`
				MetricRelabelings []interface{} `yaml:"metricRelabelings"`
			} `yaml:"serviceMonitor"`
			PrometheusRule struct {
				Enabled          bool `yaml:"enabled"`
				AdditionalLabels struct {
				} `yaml:"additionalLabels"`
				Rules []interface{} `yaml:"rules"`
			} `yaml:"prometheusRule"`
		} `yaml:"metrics"`
		Lifecycle struct {
			PreStop struct {
				Exec struct {
					Command []string `yaml:"command"`
				} `yaml:"exec"`
			} `yaml:"preStop"`
		} `yaml:"lifecycle"`
		PriorityClassName string `yaml:"priorityClassName"`
	} `yaml:"controller"`
	RevisionHistoryLimit int `yaml:"revisionHistoryLimit"`
	DefaultBackend       struct {
		Enabled bool   `yaml:"enabled"`
		Name    string `yaml:"name"`
		Image   struct {
			Registry                 string `yaml:"registry"`
			Image                    string `yaml:"image"`
			Tag                      string `yaml:"tag"`
			PullPolicy               string `yaml:"pullPolicy"`
			RunAsUser                int    `yaml:"runAsUser"`
			RunAsNonRoot             bool   `yaml:"runAsNonRoot"`
			ReadOnlyRootFilesystem   bool   `yaml:"readOnlyRootFilesystem"`
			AllowPrivilegeEscalation bool   `yaml:"allowPrivilegeEscalation"`
		} `yaml:"image"`
		ExistingPsp string `yaml:"existingPsp"`
		ExtraArgs   struct {
		} `yaml:"extraArgs"`
		ServiceAccount struct {
			Create                       bool   `yaml:"create"`
			Name                         string `yaml:"name"`
			AutomountServiceAccountToken bool   `yaml:"automountServiceAccountToken"`
		} `yaml:"serviceAccount"`
		ExtraEnvs     []interface{} `yaml:"extraEnvs"`
		Port          int           `yaml:"port"`
		LivenessProbe struct {
			FailureThreshold    int `yaml:"failureThreshold"`
			InitialDelaySeconds int `yaml:"initialDelaySeconds"`
			PeriodSeconds       int `yaml:"periodSeconds"`
			SuccessThreshold    int `yaml:"successThreshold"`
			TimeoutSeconds      int `yaml:"timeoutSeconds"`
		} `yaml:"livenessProbe"`
		ReadinessProbe struct {
			FailureThreshold    int `yaml:"failureThreshold"`
			InitialDelaySeconds int `yaml:"initialDelaySeconds"`
			PeriodSeconds       int `yaml:"periodSeconds"`
			SuccessThreshold    int `yaml:"successThreshold"`
			TimeoutSeconds      int `yaml:"timeoutSeconds"`
		} `yaml:"readinessProbe"`
		Tolerations []interface{} `yaml:"tolerations"`
		Affinity    struct {
		} `yaml:"affinity"`
		PodSecurityContext struct {
		} `yaml:"podSecurityContext"`
		ContainerSecurityContext struct {
		} `yaml:"containerSecurityContext"`
		PodLabels struct {
		} `yaml:"podLabels"`
		NodeSelector struct {
			KubernetesIoOs string `yaml:"kubernetes.io/os"`
		} `yaml:"nodeSelector"`
		PodAnnotations struct {
		} `yaml:"podAnnotations"`
		ReplicaCount int `yaml:"replicaCount"`
		MinAvailable int `yaml:"minAvailable"`
		Resources    struct {
		} `yaml:"resources"`
		ExtraVolumeMounts []interface{} `yaml:"extraVolumeMounts"`
		ExtraVolumes      []interface{} `yaml:"extraVolumes"`
		Autoscaling       struct {
			Annotations struct {
			} `yaml:"annotations"`
			Enabled                           bool `yaml:"enabled"`
			MinReplicas                       int  `yaml:"minReplicas"`
			MaxReplicas                       int  `yaml:"maxReplicas"`
			TargetCPUUtilizationPercentage    int  `yaml:"targetCPUUtilizationPercentage"`
			TargetMemoryUtilizationPercentage int  `yaml:"targetMemoryUtilizationPercentage"`
		} `yaml:"autoscaling"`
		Service struct {
			Annotations struct {
			} `yaml:"annotations"`
			ExternalIPs              []interface{} `yaml:"externalIPs"`
			LoadBalancerSourceRanges []interface{} `yaml:"loadBalancerSourceRanges"`
			ServicePort              int           `yaml:"servicePort"`
			Type                     string        `yaml:"type"`
		} `yaml:"service"`
		PriorityClassName string `yaml:"priorityClassName"`
		Labels            struct {
		} `yaml:"labels"`
	} `yaml:"defaultBackend"`
	Rbac struct {
		Create bool `yaml:"create"`
		Scope  bool `yaml:"scope"`
	} `yaml:"rbac"`
	PodSecurityPolicy struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"podSecurityPolicy"`
	ServiceAccount struct {
		Create                       bool   `yaml:"create"`
		Name                         string `yaml:"name"`
		AutomountServiceAccountToken bool   `yaml:"automountServiceAccountToken"`
		Annotations                  struct {
		} `yaml:"annotations"`
	} `yaml:"serviceAccount"`
	ImagePullSecrets []interface{} `yaml:"imagePullSecrets"`
	TCP              struct {
	} `yaml:"tcp"`
	UDP struct {
	} `yaml:"udp"`
	PortNamePrefix string      `yaml:"portNamePrefix"`
	DhParam        interface{} `yaml:"dhParam"`
}
