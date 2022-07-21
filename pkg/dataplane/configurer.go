package dataplane

import (
	"fmt"
	"time"

	"github.com/mitchellh/hashstructure"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/klog/v2"
)

// syncIngress collects all the pieces required to assemble the NGINX
// configuration file and passes the resulting data structures to the backend
// (OnUpdate) when a reload is deemed necessary.
// TODO (IMPORTANT!): This may not be the best approach. If a configuration has
// 500 Ingress and 20k endpoints, what is going to be the size of a gRPC call?
// Is this going to be a performance issue?
func (n *NGINXConfigurer) configureNGINX(pcfg *ingress.Configuration) error {
	// TODO: Implement rate limiter
	//n.syncRateLimiter.Accept()

	if n.isShuttingDown {
		return nil
	}

	// TODO: Extract servers
	/*
		n.metricCollector.SetSSLExpireTime(servers)
		n.metricCollector.SetSSLInfo(servers)
	*/

	if n.runningConfig.Equal(pcfg) {
		klog.V(3).Infof("No configuration change detected, skipping backend reload")
		return nil
	}
	// TODO: Extract hosts

	//n.metricCollector.SetHosts(hosts)
	if !n.IsDynamicConfigurationEnough(pcfg) {
		klog.InfoS("Configuration changes detected, backend reload required")

		hash, _ := hashstructure.Hash(pcfg, &hashstructure.HashOptions{
			TagName: "json",
		})

		pcfg.ConfigurationChecksum = fmt.Sprintf("%v", hash)

		err := n.OnUpdate(*pcfg)
		if err != nil {
			n.metricCollector.IncReloadErrorCount()
			n.metricCollector.ConfigSuccess(hash, false)
			klog.Errorf("Unexpected failure reloading the backend:\n%v", err)

			// TODO: Turn into a gRPC Callback with event status
			//n.recorder.Eventf(k8s.IngressPodDetails, apiv1.EventTypeWarning, "RELOAD", fmt.Sprintf("Error reloading NGINX: %v", err))
			return err
		}

		klog.InfoS("Backend successfully reloaded")
		n.metricCollector.ConfigSuccess(hash, true)
		n.metricCollector.IncReloadCount()
		// TODO: Turn into a gRPC Callback with event status
		// n.recorder.Eventf(k8s.IngressPodDetails, apiv1.EventTypeNormal, "RELOAD", "NGINX reload triggered due to a change in configuration")
	}

	isFirstSync := n.runningConfig.Equal(&ingress.Configuration{})
	if isFirstSync {
		// For the initial sync it always takes some time for NGINX to start listening
		// For large configurations it might take a while so we loop and back off
		klog.InfoS("Initial sync, sleeping for 1 second")
		time.Sleep(1 * time.Second)
	}

	retry := wait.Backoff{
		Steps:    1 + n.cfg.DynamicConfigurationRetries,
		Duration: time.Second,
		Factor:   1.3,
		Jitter:   0.1,
	}

	retriesRemaining := retry.Steps
	err := wait.ExponentialBackoff(retry, func() (bool, error) {
		err := n.configureDynamically(pcfg)
		if err == nil {
			klog.V(2).Infof("Dynamic reconfiguration succeeded.")
			return true, nil
		}
		retriesRemaining--
		if retriesRemaining > 0 {
			klog.Warningf("Dynamic reconfiguration failed (retrying; %d retries left): %v", retriesRemaining, err)
			return false, nil
		}
		klog.Warningf("Dynamic reconfiguration failed: %v", err)
		return false, err
	})
	if err != nil {
		klog.Errorf("Unexpected failure reconfiguring NGINX:\n%v", err)
		return err
	}

	ri := getRemovedIngresses(n.runningConfig, pcfg)
	re := getRemovedHosts(n.runningConfig, pcfg)
	rc := getRemovedCertificateSerialNumbers(n.runningConfig, pcfg)
	n.metricCollector.RemoveMetrics(ri, re, rc)

	n.runningConfig = pcfg

	return nil
}

// getRemovedHosts returns a list of the hostnames
// that are not associated anymore to the NGINX configuration.
func getRemovedHosts(rucfg, newcfg *ingress.Configuration) []string {
	old := sets.NewString()
	new := sets.NewString()

	for _, s := range rucfg.Servers {
		if !old.Has(s.Hostname) {
			old.Insert(s.Hostname)
		}
	}

	for _, s := range newcfg.Servers {
		if !new.Has(s.Hostname) {
			new.Insert(s.Hostname)
		}
	}

	return old.Difference(new).List()
}

func getRemovedCertificateSerialNumbers(rucfg, newcfg *ingress.Configuration) []string {
	oldCertificates := sets.NewString()
	newCertificates := sets.NewString()

	for _, server := range rucfg.Servers {
		if server.SSLCert == nil {
			continue
		}
		identifier := server.SSLCert.Identifier()
		if identifier != "" {
			if !oldCertificates.Has(identifier) {
				oldCertificates.Insert(identifier)
			}
		}
	}

	for _, server := range newcfg.Servers {
		if server.SSLCert == nil {
			continue
		}
		identifier := server.SSLCert.Identifier()
		if identifier != "" {
			if !newCertificates.Has(identifier) {
				newCertificates.Insert(identifier)
			}
		}
	}

	return oldCertificates.Difference(newCertificates).List()
}

// TODO: Maybe move those helper functions to someone else
func getRemovedIngresses(rucfg, newcfg *ingress.Configuration) []string {
	oldIngresses := sets.NewString()
	newIngresses := sets.NewString()

	for _, server := range rucfg.Servers {
		for _, location := range server.Locations {
			if location.Ingress == nil {
				continue
			}

			ingKey := k8s.MetaNamespaceKey(location.Ingress)
			if !oldIngresses.Has(ingKey) {
				oldIngresses.Insert(ingKey)
			}
		}
	}

	for _, server := range newcfg.Servers {
		for _, location := range server.Locations {
			if location.Ingress == nil {
				continue
			}

			ingKey := k8s.MetaNamespaceKey(location.Ingress)
			if !newIngresses.Has(ingKey) {
				newIngresses.Insert(ingKey)
			}
		}
	}

	return oldIngresses.Difference(newIngresses).List()
}
