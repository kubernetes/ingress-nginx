/*
Copyright 2022 The Kubernetes Authors.

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

package dataplane

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/mitchellh/hashstructure"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/ingress-nginx/internal/ingress/annotations/auth"
	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	ing_net "k8s.io/ingress-nginx/internal/net"
	"k8s.io/ingress-nginx/internal/net/ssl"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
	authfile "k8s.io/ingress-nginx/pkg/util/auth"
	"k8s.io/ingress-nginx/pkg/util/file"
	utilingress "k8s.io/ingress-nginx/pkg/util/ingress"
	ingressruntime "k8s.io/ingress-nginx/pkg/util/runtime"

	"k8s.io/klog/v2"
)

// syncIngress collects all the pieces required to assemble the NGINX
// configuration file and passes the resulting data structures to the backend
// (OnUpdate) when a reload is deemed necessary.
// It doesn't return an error, but instead sends the error to the event channel
func (n *NGINXConfigurer) syncIngress(cfg *config.TemplateConfig) {
	n.fullReconfiguration(cfg)
}

func (n *NGINXConfigurer) fullReconfiguration(cfg *config.TemplateConfig) {
	// Let's configure just once per time.
	n.configureLock.Lock()
	defer n.configureLock.Unlock()

	n.metricCollector.SetSSLExpireTime(cfg.Servers)
	n.metricCollector.SetSSLInfo(cfg.Servers)

	klog.V(3).Infof("New reconfiguration requested. Old config timestamp: %v New config timestamp: %v", n.templateConfig.GeneratedTime, cfg.GeneratedTime)

	if n.templateConfig.GeneratedTime >= cfg.GeneratedTime {
		klog.V(3).Infof("No configuration change detected, skipping backend reload")
		return
	}

	// TODO: Get metrics from the right place
	// n.metricCollector.SetHosts(hosts)
	newcfg := configFromTemplate(cfg)
	oldcfg := configFromTemplate(n.templateConfig)

	if cfg.DefaultSSLCertificate != nil {
		if (n.templateConfig.DefaultSSLCertificate == nil) || (n.templateConfig.DefaultSSLCertificate.PemSHA != cfg.DefaultSSLCertificate.PemSHA) {
			name := cfg.DefaultSSLCertificate.Name
			if name == "" {
				name = ssl.FakeCertificateName
			}
			if _, err := ssl.StoreSSLCertOnDisk(name, cfg.DefaultSSLCertificate); err != nil {
				klog.ErrorS(err, "failed to create default certificate file")
				return
			}
		}
	}

	if cfg.DHParamFile != "" && cfg.DHParamContent != nil {
		if err := ssl.AddOrUpdateDHParam(cfg.DHParamFile, cfg.DHParamContent); err != nil {
			klog.ErrorS(err, "failed to write dh param file")
			return
		}
	}

	var configMapChange bool
	if n.templateConfig != nil {
		if n.templateConfig.Cfg.Checksum != cfg.Cfg.Checksum {
			klog.V(3).Infof("Configmap has changed. Will trigger a reload. Old checksum: %s New checksum: %s",
				n.templateConfig.Cfg.Checksum, cfg.Cfg.Checksum)
			configMapChange = true
		}
	}

	certChange := checkAndWriteDeltaCertificates(cfg, n.templateConfig)
	authChange := checkAndWriteAuthSecrets(cfg, n.templateConfig)

	if authChange || certChange || configMapChange || !utilingress.IsDynamicConfigurationEnough(newcfg, oldcfg) {
		klog.InfoS("Configuration changes detected, backend reload required")

		hash, err := hashstructure.Hash(newcfg, &hashstructure.HashOptions{
			TagName: "json",
		})
		if err != nil {
			klog.Warningf("error generating config hash: %s", err)
			return
		}

		newcfg.ConfigurationChecksum = fmt.Sprintf("%v", hash)

		err = n.updateConfiguration(cfg)
		if err != nil {
			n.metricCollector.IncReloadErrorCount()
			n.metricCollector.ConfigSuccess(hash, false)
			klog.Errorf("Unexpected failure reloading the backend:\n%v", err)
			n.GRPCClient.EventCh <- newEventMessage(apiv1.EventTypeWarning, "RELOAD", fmt.Sprintf("Error reloading NGINX: %v", err))
			return

		}

		klog.InfoS("Backend successfully reloaded")
		n.metricCollector.ConfigSuccess(hash, true)
		n.metricCollector.IncReloadCount()

		n.GRPCClient.EventCh <- newEventMessage(apiv1.EventTypeNormal, "RELOAD", "NGINX reload triggered due to a change in configuration")
	}

	isFirstSync := oldcfg.Equal(&ingress.Configuration{})
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
		err := utilingress.ConfigureDynamically(newcfg, oldcfg)
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
		n.GRPCClient.EventCh <- newEventMessage("Error", "RELOAD", fmt.Sprintf("Unexpected failure reconfiguring NGINX:\n%v", err))
		return
	}

	ri := utilingress.GetRemovedIngresses(oldcfg, newcfg)
	re := utilingress.GetRemovedHosts(oldcfg, newcfg)
	rc := utilingress.GetRemovedCertificateSerialNumbers(oldcfg, newcfg)

	n.metricCollector.RemoveMetrics(ri, re, rc)

	n.templateConfig = cfg

}

// TODO: move to a better place
type certOperation struct {
	cert      config.CertificateFile
	operation CertOperationType
}

type CertOperationType string

const (
	certAdd    CertOperationType = "ADD"
	certRemove CertOperationType = "REMOVE"
)

// TODO: UNIT TEST!!! And e2e tests!
func checkAndWriteAuthSecrets(newtmpl *config.TemplateConfig, oldtmpl *config.TemplateConfig) bool {
	var changed bool

	tempAuthOld := make(map[string]auth.Config)
	tempAuthNew := make(map[string]auth.Config)

	// Get added auth secrets:
	for _, v := range newtmpl.Servers {
		for _, location := range v.Locations {
			if location.BasicDigestAuth.FileSHA != "" {
				tempAuthNew[location.BasicDigestAuth.File] = location.BasicDigestAuth
			}
		}
	}

	// Get old auth secrets:
	for _, v := range oldtmpl.Servers {
		for _, location := range v.Locations {
			if location.BasicDigestAuth.FileSHA != "" {
				tempAuthOld[location.BasicDigestAuth.File] = location.BasicDigestAuth
			}
		}
	}

	// Check for removed secrets/files.
	for authFile := range tempAuthOld {
		if _, ok := tempAuthNew[authFile]; !ok {
			changed = true
			if err := os.Remove(authFile); err != nil && !os.IsNotExist(err) {
				klog.Warningf("failed removing old auth file %s: %s", authFile, err)
				continue
			}
		}
	}

	// Check on newMap if the files already exists and SHA matches, otherwise create/update
	for newAuthFile, newAuthConfig := range tempAuthNew {
		// If the auth secret didn't existed or existed but is not equal, then rewrite
		if oldAuthConfig, ok := tempAuthOld[newAuthFile]; !ok || oldAuthConfig.FileSHA != newAuthConfig.FileSHA {
			changed = true
			if err := authfile.WriteSecretFile(newAuthFile, newAuthConfig.SecretContent); err != nil {
				klog.Warningf("failed adding/updating auth file %s: %s", newAuthFile, err)
				continue
			}
		}
	}

	return changed
}

// TODO: UNIT TEST!!! And e2e tests!
func checkAndWriteDeltaCertificates(newtmpl *config.TemplateConfig, oldtmpl *config.TemplateConfig) bool {
	var changed bool

	tempMap := make(map[string]certOperation)
	// Get added certificates:
	for k, v := range newtmpl.PersistedCertificates {
		kold, ok := oldtmpl.PersistedCertificates[k]
		// if non existent on the old one, add to the tempMap as a newCert
		if !ok {
			// format of added file will be +;Filename being ; a separator a + the create operation
			tempMap[k] = certOperation{
				operation: certAdd,
				cert:      v,
			}
		} else {
			// if existent in the old map, let's see if it changed checking the checksum
			if v.Checksum != kold.Checksum {
				tempMap[k] = certOperation{
					operation: certAdd,
					cert:      v,
				}
			}
		}
	}

	// Now remove old certificates
	for k := range oldtmpl.PersistedCertificates {
		if _, ok := newtmpl.PersistedCertificates[k]; !ok {
			// if NOK, add on the tempMap an operation to remove it
			// removal does not need the bytes or the SHA
			tempMap[k] = certOperation{
				operation: certRemove,
				cert:      config.CertificateFile{},
			}
		}
	}

	// If so, let's do the required operations and mark the reload as required
	if len(tempMap) > 0 {
		changed = true
		for k, v := range tempMap {
			if v.operation == certRemove {
				klog.Infof("removing  cert %s", k)
				if err := os.Remove(k); err != nil {
					klog.Warningf("failed removing old certificate %s: %s", k, err)
				}
				continue
			}
			klog.Infof("adding/updating cert %s", k)
			err := os.WriteFile(k, []byte(v.cert.Content), file.ReadWriteByUser)
			if err != nil {
				klog.Errorf("could not create PEM certificate file %v: %v", k, err)
				continue
			}
		}
	}
	return changed
}

// TODO: Add unit test
// configFromTemplate converts a TemplateConfig into an ingress.Configuration
func configFromTemplate(tmpl *config.TemplateConfig) *ingress.Configuration {
	if tmpl == nil {
		return &ingress.Configuration{}
	}

	return &ingress.Configuration{
		Backends:     tmpl.Backends,
		Servers:      tmpl.Servers,
		TCPEndpoints: tmpl.TCPBackends,
		UDPEndpoints: tmpl.UDPBackends,
	}
}

func (n *NGINXConfigurer) updateConfiguration(cfg *config.TemplateConfig) error {
	err := createOpentracingCfg(cfg.Cfg)
	if err != nil {
		return err
	}

	cfg.Cfg.SSLDHParam = cfg.DHParamFile
	cfg.BacklogSize = ingressruntime.SysctlSomaxconn()
	cfg.Cfg.Resolver = n.resolver
	cfg.Cfg.DisableIpv6DNS = !ing_net.IsIPv6Enabled()

	content, err := n.t.Write(*cfg)
	if err != nil {
		return err
	}

	err = n.testTemplate(content)
	if err != nil {
		return err
	}

	if klog.V(2).Enabled() {
		src, _ := os.ReadFile(cfgPath)
		if !bytes.Equal(src, content) {
			tmpfile, err := os.CreateTemp("", "new-nginx-cfg")
			if err != nil {
				return err
			}
			defer tmpfile.Close()
			err = os.WriteFile(tmpfile.Name(), content, file.ReadWriteByUser)
			if err != nil {
				return err
			}

			diffOutput, err := exec.Command("diff", "-I", "'# Configuration.*'", "-u", cfgPath, tmpfile.Name()).CombinedOutput()
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					ws := exitError.Sys().(syscall.WaitStatus)
					if ws.ExitStatus() == 2 {
						klog.Warningf("Failed to executing diff command: %v", err)
					}
				}
			}

			klog.InfoS("NGINX configuration change", "diff", string(diffOutput))

			// we do not defer the deletion of temp files in order
			// to keep them around for inspection in case of error
			os.Remove(tmpfile.Name())
		}
	}

	err = os.WriteFile(cfgPath, content, file.ReadWriteByUser)
	if err != nil {
		return err
	}

	o, err := n.command.ExecCommand("-s", "reload").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v\n%v", err, string(o))
	}

	return nil
}
