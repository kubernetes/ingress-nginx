/*
Copyright 2019 The Kubernetes Authors.

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

package request

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	discoveryv1client "k8s.io/client-go/kubernetes/typed/discovery/v1"
	typednetworking "k8s.io/client-go/kubernetes/typed/networking/v1"

	"k8s.io/ingress-nginx/cmd/plugin/util"
)

// ChoosePod finds a pod either by deployment or by name
func ChoosePod(flags *genericclioptions.ConfigFlags, podName string, deployment string, selector string) (apiv1.Pod, error) {
	if podName != "" {
		return GetNamedPod(flags, podName)
	}

	if selector != "" {
		return GetLabeledPod(flags, selector)
	}

	return GetDeploymentPod(flags, deployment)
}

// GetNamedPod finds a pod with the given name
func GetNamedPod(flags *genericclioptions.ConfigFlags, name string) (apiv1.Pod, error) {
	allPods, err := getPods(flags)
	if err != nil {
		return apiv1.Pod{}, err
	}

	for _, pod := range allPods {
		if pod.Name == name {
			return pod, nil
		}
	}

	return apiv1.Pod{}, fmt.Errorf("pod %v not found in namespace %v", name, util.GetNamespace(flags))
}

// GetDeploymentPod finds a pod from a given deployment
func GetDeploymentPod(flags *genericclioptions.ConfigFlags, deployment string) (apiv1.Pod, error) {
	ings, err := getDeploymentPods(flags, deployment)
	if err != nil {
		return apiv1.Pod{}, err
	}

	if len(ings) == 0 {
		return apiv1.Pod{}, fmt.Errorf("no pods for deployment %v found in namespace %v", deployment, util.GetNamespace(flags))
	}

	return ings[0], nil
}

// GetLabeledPod finds a pod from a given label
func GetLabeledPod(flags *genericclioptions.ConfigFlags, label string) (apiv1.Pod, error) {
	ings, err := getLabeledPods(flags, label)
	if err != nil {
		return apiv1.Pod{}, err
	}

	if len(ings) == 0 {
		return apiv1.Pod{}, fmt.Errorf("no pods for label selector %v found in namespace %v", label, util.GetNamespace(flags))
	}

	return ings[0], nil
}

// GetDeployments returns an array of Deployments
func GetDeployments(flags *genericclioptions.ConfigFlags, namespace string) ([]appsv1.Deployment, error) {
	rawConfig, err := flags.ToRESTConfig()
	if err != nil {
		return make([]appsv1.Deployment, 0), err
	}

	api, err := appsv1client.NewForConfig(rawConfig)
	if err != nil {
		return make([]appsv1.Deployment, 0), err
	}

	deployments, err := api.Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return make([]appsv1.Deployment, 0), err
	}

	return deployments.Items, nil
}

// GetIngressDefinitions returns an array of Ingress resource definitions
func GetIngressDefinitions(flags *genericclioptions.ConfigFlags, namespace string) ([]networking.Ingress, error) {
	rawConfig, err := flags.ToRESTConfig()
	if err != nil {
		return make([]networking.Ingress, 0), err
	}

	api, err := typednetworking.NewForConfig(rawConfig)
	if err != nil {
		return make([]networking.Ingress, 0), err
	}

	pods, err := api.Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return make([]networking.Ingress, 0), err
	}

	return pods.Items, nil
}

// GetNumEndpoints counts the number of endpointslices adresses for the service with the given name
func GetNumEndpoints(flags *genericclioptions.ConfigFlags, namespace string, serviceName string) (*int, error) {
	epss, err := GetEndpointSlicesByName(flags, namespace, serviceName)
	if err != nil {
		return nil, err
	}

	if len(epss) == 0 {
		return nil, nil
	}

	ret := 0
	for _, eps := range epss {
		for _, ep := range eps.Endpoints {
			ret += len(ep.Addresses)
		}
	}
	return &ret, nil
}

// GetEndpointSlicesByName returns the endpointSlices for the service with the given name
func GetEndpointSlicesByName(flags *genericclioptions.ConfigFlags, namespace string, name string) ([]discoveryv1.EndpointSlice, error) {
	allEndpointsSlices, err := getEndpointSlices(flags, namespace)
	if err != nil {
		return nil, err
	}
	var eps []discoveryv1.EndpointSlice
	for _, slice := range allEndpointsSlices {
		if svcName, ok := slice.ObjectMeta.GetLabels()[discoveryv1.LabelServiceName]; ok {
			if svcName == name {
				eps = append(eps, slice)
			}
		}
	}

	return eps, nil
}

var endpointSlicesCache = make(map[string]*[]discoveryv1.EndpointSlice)

func getEndpointSlices(flags *genericclioptions.ConfigFlags, namespace string) ([]discoveryv1.EndpointSlice, error) {
	cachedEndpointSlices, ok := endpointSlicesCache[namespace]

	if ok {
		return *cachedEndpointSlices, nil
	}

	if namespace != "" {
		tryAllNamespacesEndpointSlicesCache(flags)
	}

	cachedEndpointSlices = tryFilteringEndpointSlicesFromAllNamespacesCache(flags, namespace)

	if cachedEndpointSlices != nil {
		return *cachedEndpointSlices, nil
	}

	rawConfig, err := flags.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	api, err := discoveryv1client.NewForConfig(rawConfig)
	if err != nil {
		return nil, err
	}
	endpointSlicesList, err := api.EndpointSlices(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	endpointSlices := endpointSlicesList.Items

	endpointSlicesCache[namespace] = &endpointSlices
	return endpointSlices, nil
}

func tryAllNamespacesEndpointSlicesCache(flags *genericclioptions.ConfigFlags) {
	_, ok := endpointSlicesCache[""]
	if !ok {
		_, err := getEndpointSlices(flags, "")
		if err != nil {
			endpointSlicesCache[""] = nil
		}
	}
}

func tryFilteringEndpointSlicesFromAllNamespacesCache(flags *genericclioptions.ConfigFlags, namespace string) *[]discoveryv1.EndpointSlice {
	allEndpointSlices := endpointSlicesCache[""]
	if allEndpointSlices != nil {
		endpointSlices := make([]discoveryv1.EndpointSlice, 0)
		for _, slice := range *allEndpointSlices {
			if slice.Namespace == namespace {
				endpointSlices = append(endpointSlices, slice)
			}
		}
		endpointSlicesCache[namespace] = &endpointSlices
		return &endpointSlices
	}
	return nil
}

// GetServiceByName finds and returns the service definition with the given name
func GetServiceByName(flags *genericclioptions.ConfigFlags, name string, services *[]apiv1.Service) (apiv1.Service, error) {
	if services == nil {
		servicesArray, err := getServices(flags)
		if err != nil {
			return apiv1.Service{}, err
		}
		services = &servicesArray
	}

	for _, svc := range *services {
		if svc.Name == name {
			return svc, nil
		}
	}

	return apiv1.Service{}, fmt.Errorf("could not find service %v in namespace %v", name, util.GetNamespace(flags))
}

func getPods(flags *genericclioptions.ConfigFlags) ([]apiv1.Pod, error) {
	namespace := util.GetNamespace(flags)

	rawConfig, err := flags.ToRESTConfig()
	if err != nil {
		return make([]apiv1.Pod, 0), err
	}

	api, err := corev1.NewForConfig(rawConfig)
	if err != nil {
		return make([]apiv1.Pod, 0), err
	}

	pods, err := api.Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return make([]apiv1.Pod, 0), err
	}

	return pods.Items, nil
}

func getLabeledPods(flags *genericclioptions.ConfigFlags, label string) ([]apiv1.Pod, error) {
	namespace := util.GetNamespace(flags)

	rawConfig, err := flags.ToRESTConfig()
	if err != nil {
		return make([]apiv1.Pod, 0), err
	}

	api, err := corev1.NewForConfig(rawConfig)
	if err != nil {
		return make([]apiv1.Pod, 0), err
	}

	pods, err := api.Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: label,
	})

	if err != nil {
		return make([]apiv1.Pod, 0), err
	}

	return pods.Items, nil
}

func getDeploymentPods(flags *genericclioptions.ConfigFlags, deployment string) ([]apiv1.Pod, error) {
	pods, err := getPods(flags)
	if err != nil {
		return make([]apiv1.Pod, 0), err
	}

	ingressPods := make([]apiv1.Pod, 0)
	for _, pod := range pods {
		if util.PodInDeployment(pod, deployment) {
			ingressPods = append(ingressPods, pod)
		}
	}

	return ingressPods, nil
}

func getServices(flags *genericclioptions.ConfigFlags) ([]apiv1.Service, error) {
	namespace := util.GetNamespace(flags)

	rawConfig, err := flags.ToRESTConfig()
	if err != nil {
		return make([]apiv1.Service, 0), err
	}

	api, err := corev1.NewForConfig(rawConfig)
	if err != nil {
		return make([]apiv1.Service, 0), err
	}

	services, err := api.Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return make([]apiv1.Service, 0), err
	}

	return services.Items, nil

}
