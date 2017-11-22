package controller

import (
	"strconv"
	"strings"

	"github.com/golang/glog"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations/healthcheck"
	"k8s.io/ingress-nginx/internal/k8s"
)

func (n *NGINXController) getStreamServices(configmapName string, proto apiv1.Protocol) []ingress.L4Service {
	glog.V(3).Infof("obtaining information about stream services of type %v located in configmap %v", proto, configmapName)
	if configmapName == "" {
		// no configmap configured
		return []ingress.L4Service{}
	}

	_, _, err := k8s.ParseNameNS(configmapName)
	if err != nil {
		glog.Errorf("unexpected error reading configmap %v: %v", configmapName, err)
		return []ingress.L4Service{}
	}

	configmap, err := n.storeLister.GetConfigMap(configmapName)
	if err != nil {
		glog.Errorf("unexpected error reading configmap %v: %v", configmapName, err)
		return []ingress.L4Service{}
	}

	var svcs []ingress.L4Service
	var svcProxyProtocol ingress.ProxyProtocol
	// k -> port to expose
	// v -> <namespace>/<service name>:<port from service to be used>
	for k, v := range configmap.Data {
		externalPort, err := strconv.Atoi(k)
		if err != nil {
			glog.Warningf("%v is not valid as a TCP/UDP port", k)
			continue
		}

		rp := []int{
			n.cfg.ListenPorts.HTTP,
			n.cfg.ListenPorts.HTTPS,
			n.cfg.ListenPorts.SSLProxy,
			n.cfg.ListenPorts.Status,
			n.cfg.ListenPorts.Health,
			n.cfg.ListenPorts.Default,
		}

		if intInSlice(externalPort, rp) {
			glog.Warningf("port %v cannot be used for TCP or UDP services. It is reserved for the Ingress controller", k)
			continue
		}

		nsSvcPort := strings.Split(v, ":")
		if len(nsSvcPort) < 2 {
			glog.Warningf("invalid format (namespace/name:port:[PROXY]:[PROXY]) '%v'", k)
			continue
		}

		nsName := nsSvcPort[0]
		svcPort := nsSvcPort[1]
		svcProxyProtocol.Decode = false
		svcProxyProtocol.Encode = false

		// Proxy protocol is possible if the service is TCP
		if len(nsSvcPort) >= 3 && proto == apiv1.ProtocolTCP {
			if len(nsSvcPort) >= 3 && strings.ToUpper(nsSvcPort[2]) == "PROXY" {
				svcProxyProtocol.Decode = true
			}
			if len(nsSvcPort) == 4 && strings.ToUpper(nsSvcPort[3]) == "PROXY" {
				svcProxyProtocol.Encode = true
			}
		}

		svcNs, svcName, err := k8s.ParseNameNS(nsName)
		if err != nil {
			glog.Warningf("%v", err)
			continue
		}

		svc, err := n.storeLister.GetService(nsName)
		if err != nil {
			glog.Warningf("error getting service %v: %v", nsName, err)
			continue
		}

		var endps []ingress.Endpoint
		targetPort, err := strconv.Atoi(svcPort)
		if err != nil {
			glog.V(3).Infof("searching service %v endpoints using the name '%v'", svcNs, svcName, svcPort)
			for _, sp := range svc.Spec.Ports {
				if sp.Name == svcPort {
					if sp.Protocol == proto {
						endps = n.getEndpoints(svc, &sp, proto, &healthcheck.Config{})
						break
					}
				}
			}
		} else {
			// we need to use the TargetPort (where the endpoints are running)
			glog.V(3).Infof("searching service %v/%v endpoints using the target port '%v'", svcNs, svcName, targetPort)
			for _, sp := range svc.Spec.Ports {
				if sp.Port == int32(targetPort) {
					if sp.Protocol == proto {
						endps = n.getEndpoints(svc, &sp, proto, &healthcheck.Config{})
						break
					}
				}
			}
		}

		// stream services cannot contain empty upstreams and there is no
		// default backend equivalent
		if len(endps) == 0 {
			glog.Warningf("service %v/%v does not have any active endpoints for port %v and protocol %v", svcNs, svcName, svcPort, proto)
			continue
		}

		svcs = append(svcs, ingress.L4Service{
			Port: externalPort,
			Backend: ingress.L4Backend{
				Name:          svcName,
				Namespace:     svcNs,
				Port:          intstr.FromString(svcPort),
				Protocol:      proto,
				ProxyProtocol: svcProxyProtocol,
			},
			Endpoints: endps,
		})
	}

	return svcs
}
