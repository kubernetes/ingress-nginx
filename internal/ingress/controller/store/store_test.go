/*
Copyright 2017 The Kubernetes Authors.

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

package store

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/eapache/channels"
	extensions "k8s.io/api/extensions/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/cache"

	"encoding/base64"
	"io/ioutil"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/ingress-nginx/internal/file"
	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

func TestStore(t *testing.T) {
	pod := &k8s.PodInfo{
		Name:      "testpod",
		Namespace: v1.NamespaceDefault,
		Labels: map[string]string{
			"pod-template-hash": "1234",
		},
	}

	clientSet := fake.NewSimpleClientset()

	t.Run("should return an error searching for non existing objects", func(t *testing.T) {
		ns := createNamespace(clientSet, t)
		defer deleteNamespace(ns, clientSet, t)
		cm := createConfigMap(clientSet, ns, t)
		defer deleteConfigMap(cm, ns, clientSet, t)

		stopCh := make(chan struct{})
		updateCh := channels.NewRingChannel(1024)

		go func(ch *channels.RingChannel) {
			for {
				<-ch.Out()
			}
		}(updateCh)

		fs := newFS(t)
		storer := New(true,
			ns,
			fmt.Sprintf("%v/config", ns),
			fmt.Sprintf("%v/tcp", ns),
			fmt.Sprintf("%v/udp", ns),
			"",
			10*time.Minute,
			clientSet,
			fs,
			updateCh,
			false,
			pod,
			false)

		storer.Run(stopCh)

		key := fmt.Sprintf("%v/anything", ns)
		ls, err := storer.GetLocalSSLCert(key)
		if err == nil {
			t.Errorf("expected an error but none returned")
		}
		if ls != nil {
			t.Errorf("expected an Ingres but none returned")
		}

		s, err := storer.GetSecret(key)
		if err == nil {
			t.Errorf("expected an error but none returned")
		}
		if s != nil {
			t.Errorf("expected an Ingres but none returned")
		}

		svc, err := storer.GetService(key)
		if err == nil {
			t.Errorf("expected an error but none returned")
		}
		if svc != nil {
			t.Errorf("expected an Ingres but none returned")
		}
	})

	t.Run("should return one event for add, update and delete of ingress", func(t *testing.T) {
		ns := createNamespace(clientSet, t)
		defer deleteNamespace(ns, clientSet, t)
		cm := createConfigMap(clientSet, ns, t)
		defer deleteConfigMap(cm, ns, clientSet, t)

		stopCh := make(chan struct{})
		updateCh := channels.NewRingChannel(1024)

		var add uint64
		var upd uint64
		var del uint64

		go func(ch *channels.RingChannel) {
			for {
				evt, ok := <-ch.Out()
				if !ok {
					return
				}

				e := evt.(Event)
				if e.Obj == nil {
					continue
				}
				if _, ok := e.Obj.(*extensions.Ingress); !ok {
					continue
				}

				switch e.Type {
				case CreateEvent:
					atomic.AddUint64(&add, 1)
				case UpdateEvent:
					atomic.AddUint64(&upd, 1)
				case DeleteEvent:
					atomic.AddUint64(&del, 1)
				}
			}
		}(updateCh)

		fs := newFS(t)
		storer := New(true,
			ns,
			fmt.Sprintf("%v/config", ns),
			fmt.Sprintf("%v/tcp", ns),
			fmt.Sprintf("%v/udp", ns),
			"",
			10*time.Minute,
			clientSet,
			fs,
			updateCh,
			false,
			pod,
			false)

		storer.Run(stopCh)

		ing := ensureIngress(&extensions.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dummy",
				Namespace: ns,
				SelfLink:  fmt.Sprintf("/apis/extensions/v1beta1/namespaces/%s/ingresses/dummy", ns),
			},
			Spec: extensions.IngressSpec{
				Rules: []extensions.IngressRule{
					{
						Host: "dummy",
						IngressRuleValue: extensions.IngressRuleValue{
							HTTP: &extensions.HTTPIngressRuleValue{
								Paths: []extensions.HTTPIngressPath{
									{
										Path: "/",
										Backend: extensions.IngressBackend{
											ServiceName: "http-svc",
											ServicePort: intstr.FromInt(80),
										},
									},
								},
							},
						},
					},
				},
			},
		}, clientSet, t)

		err := framework.WaitForIngressInNamespace(clientSet, ns, ing.Name)
		if err != nil {
			t.Errorf("error waiting for secret: %v", err)
		}
		time.Sleep(1 * time.Second)

		// create an invalid ingress (different class)
		invalidIngress := ensureIngress(&extensions.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "custom-class",
				SelfLink:  fmt.Sprintf("/apis/extensions/v1beta1/namespaces/%s/ingresses/custom-class", ns),
				Namespace: ns,
				Annotations: map[string]string{
					"kubernetes.io/ingress.class": "something",
				},
			},
			Spec: extensions.IngressSpec{
				Rules: []extensions.IngressRule{
					{
						Host: "dummy",
						IngressRuleValue: extensions.IngressRuleValue{
							HTTP: &extensions.HTTPIngressRuleValue{
								Paths: []extensions.HTTPIngressPath{
									{
										Path: "/",
										Backend: extensions.IngressBackend{
											ServiceName: "http-svc",
											ServicePort: intstr.FromInt(80),
										},
									},
								},
							},
						},
					},
				},
			},
		}, clientSet, t)
		defer deleteIngress(invalidIngress, clientSet, t)

		ni := ing.DeepCopy()
		ni.Spec.Rules[0].Host = "update-dummy"
		_ = ensureIngress(ni, clientSet, t)
		if err != nil {
			t.Errorf("error creating ingress: %v", err)
		}
		// Secret takes a bit to update
		time.Sleep(3 * time.Second)

		err = clientSet.Extensions().Ingresses(ni.Namespace).Delete(ni.Name, &metav1.DeleteOptions{})
		if err != nil {
			t.Errorf("error creating ingress: %v", err)
		}

		err = framework.WaitForNoIngressInNamespace(clientSet, ni.Namespace, ni.Name)
		if err != nil {
			t.Errorf("error waiting for secret: %v", err)
		}
		time.Sleep(1 * time.Second)

		if atomic.LoadUint64(&add) != 1 {
			t.Errorf("expected 1 event of type Create but %v occurred", add)
		}
		if atomic.LoadUint64(&upd) != 1 {
			t.Errorf("expected 1 event of type Update but %v occurred", upd)
		}
		if atomic.LoadUint64(&del) != 1 {
			t.Errorf("expected 1 event of type Delete but %v occurred", del)
		}
	})

	t.Run("should not receive updates for ingress with invalid class", func(t *testing.T) {
		ns := createNamespace(clientSet, t)
		defer deleteNamespace(ns, clientSet, t)
		cm := createConfigMap(clientSet, ns, t)
		defer deleteConfigMap(cm, ns, clientSet, t)

		stopCh := make(chan struct{})
		updateCh := channels.NewRingChannel(1024)

		var add uint64
		var upd uint64
		var del uint64

		go func(ch *channels.RingChannel) {
			for {
				evt, ok := <-ch.Out()
				if !ok {
					return
				}

				e := evt.(Event)
				if e.Obj == nil {
					continue
				}
				if _, ok := e.Obj.(*extensions.Ingress); !ok {
					continue
				}

				switch e.Type {
				case CreateEvent:
					atomic.AddUint64(&add, 1)
				case UpdateEvent:
					atomic.AddUint64(&upd, 1)
				case DeleteEvent:
					atomic.AddUint64(&del, 1)
				}
			}
		}(updateCh)

		fs := newFS(t)
		storer := New(true,
			ns,
			fmt.Sprintf("%v/config", ns),
			fmt.Sprintf("%v/tcp", ns),
			fmt.Sprintf("%v/udp", ns),
			"",
			10*time.Minute,
			clientSet,
			fs,
			updateCh,
			false,
			pod,
			false)

		storer.Run(stopCh)

		// create an invalid ingress (different class)
		invalidIngress := ensureIngress(&extensions.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "custom-class",
				SelfLink:  fmt.Sprintf("/apis/extensions/v1beta1/namespaces/%s/ingresses/custom-class", ns),
				Namespace: ns,
				Annotations: map[string]string{
					"kubernetes.io/ingress.class": "something",
				},
			},
			Spec: extensions.IngressSpec{
				Rules: []extensions.IngressRule{
					{
						Host: "dummy",
						IngressRuleValue: extensions.IngressRuleValue{
							HTTP: &extensions.HTTPIngressRuleValue{
								Paths: []extensions.HTTPIngressPath{
									{
										Path: "/",
										Backend: extensions.IngressBackend{
											ServiceName: "http-svc",
											ServicePort: intstr.FromInt(80),
										},
									},
								},
							},
						},
					},
				},
			},
		}, clientSet, t)
		err := framework.WaitForIngressInNamespace(clientSet, ns, invalidIngress.Name)
		if err != nil {
			t.Errorf("error waiting for ingress: %v", err)
		}
		time.Sleep(1 * time.Second)

		invalidIngressUpdated := invalidIngress.DeepCopy()
		invalidIngressUpdated.Spec.Rules[0].Host = "update-dummy"
		_ = ensureIngress(invalidIngressUpdated, clientSet, t)
		if err != nil {
			t.Errorf("error creating ingress: %v", err)
		}
		// Secret takes a bit to update
		time.Sleep(3 * time.Second)

		if atomic.LoadUint64(&add) != 0 {
			t.Errorf("expected 0 event of type Create but %v occurred", add)
		}
		if atomic.LoadUint64(&upd) != 0 {
			t.Errorf("expected 0 event of type Update but %v occurred", upd)
		}
		if atomic.LoadUint64(&del) != 0 {
			t.Errorf("expected 0 event of type Delete but %v occurred", del)
		}
	})

	t.Run("should not receive events from secret not referenced from ingress", func(t *testing.T) {
		ns := createNamespace(clientSet, t)
		defer deleteNamespace(ns, clientSet, t)
		cm := createConfigMap(clientSet, ns, t)
		defer deleteConfigMap(cm, ns, clientSet, t)

		stopCh := make(chan struct{})
		updateCh := channels.NewRingChannel(1024)

		var add uint64
		var upd uint64
		var del uint64

		go func(ch *channels.RingChannel) {
			for {
				evt, ok := <-ch.Out()
				if !ok {
					return
				}

				e := evt.(Event)
				if e.Obj == nil {
					continue
				}
				switch e.Type {
				case CreateEvent:
					atomic.AddUint64(&add, 1)
				case UpdateEvent:
					atomic.AddUint64(&upd, 1)
				case DeleteEvent:
					atomic.AddUint64(&del, 1)
				}
			}
		}(updateCh)

		fs := newFS(t)
		storer := New(true,
			ns,
			fmt.Sprintf("%v/config", ns),
			fmt.Sprintf("%v/tcp", ns),
			fmt.Sprintf("%v/udp", ns),
			"",
			10*time.Minute,
			clientSet,
			fs,
			updateCh,
			false,
			pod,
			false)

		storer.Run(stopCh)

		secretName := "not-referenced"
		_, err := framework.CreateIngressTLSSecret(clientSet, []string{"foo"}, secretName, ns)
		if err != nil {
			t.Errorf("error creating secret: %v", err)
		}

		err = framework.WaitForSecretInNamespace(clientSet, ns, secretName)
		if err != nil {
			t.Errorf("error waiting for secret: %v", err)
		}

		if atomic.LoadUint64(&add) != 0 {
			t.Errorf("expected 0 events of type Create but %v occurred", add)
		}
		if atomic.LoadUint64(&upd) != 0 {
			t.Errorf("expected 0 events of type Update but %v occurred", upd)
		}
		if atomic.LoadUint64(&del) != 0 {
			t.Errorf("expected 0 events of type Delete but %v occurred", del)
		}

		err = clientSet.CoreV1().Secrets(ns).Delete(secretName, &metav1.DeleteOptions{})
		if err != nil {
			t.Errorf("error deleting secret: %v", err)
		}

		time.Sleep(1 * time.Second)

		if atomic.LoadUint64(&add) != 0 {
			t.Errorf("expected 0 events of type Create but %v occurred", add)
		}
		if atomic.LoadUint64(&upd) != 0 {
			t.Errorf("expected 0 events of type Update but %v occurred", upd)
		}
		if atomic.LoadUint64(&del) != 0 {
			t.Errorf("expected 0 events of type Delete but %v occurred", del)
		}
	})

	t.Run("should receive events from secret referenced from ingress", func(t *testing.T) {
		ns := createNamespace(clientSet, t)
		defer deleteNamespace(ns, clientSet, t)
		cm := createConfigMap(clientSet, ns, t)
		defer deleteConfigMap(cm, ns, clientSet, t)

		stopCh := make(chan struct{})
		updateCh := channels.NewRingChannel(1024)

		var add uint64
		var upd uint64
		var del uint64

		go func(ch *channels.RingChannel) {
			for {
				evt, ok := <-ch.Out()
				if !ok {
					return
				}

				e := evt.(Event)
				if e.Obj == nil {
					continue
				}
				switch e.Type {
				case CreateEvent:
					atomic.AddUint64(&add, 1)
				case UpdateEvent:
					atomic.AddUint64(&upd, 1)
				case DeleteEvent:
					atomic.AddUint64(&del, 1)
				}
			}
		}(updateCh)

		fs := newFS(t)
		storer := New(true,
			ns,
			fmt.Sprintf("%v/config", ns),
			fmt.Sprintf("%v/tcp", ns),
			fmt.Sprintf("%v/udp", ns),
			"",
			10*time.Minute,
			clientSet,
			fs,
			updateCh,
			false,
			pod,
			false)

		storer.Run(stopCh)

		ingressName := "ingress-with-secret"
		secretName := "referenced"

		ing := ensureIngress(&extensions.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ingressName,
				Namespace: ns,
				SelfLink:  fmt.Sprintf("/apis/extensions/v1beta1/namespaces/%s/ingresses/%s", ns, ingressName),
			},
			Spec: extensions.IngressSpec{
				TLS: []extensions.IngressTLS{
					{
						SecretName: secretName,
					},
				},
				Backend: &extensions.IngressBackend{
					ServiceName: "http-svc",
					ServicePort: intstr.FromInt(80),
				},
			},
		}, clientSet, t)
		defer deleteIngress(ing, clientSet, t)

		err := framework.WaitForIngressInNamespace(clientSet, ns, ingressName)
		if err != nil {
			t.Errorf("error waiting for secret: %v", err)
		}

		_, err = framework.CreateIngressTLSSecret(clientSet, []string{"foo"}, secretName, ns)
		if err != nil {
			t.Errorf("error creating secret: %v", err)
		}

		err = framework.WaitForSecretInNamespace(clientSet, ns, secretName)
		if err != nil {
			t.Errorf("error waiting for secret: %v", err)
		}

		// take into account secret sync
		time.Sleep(3 * time.Second)

		if atomic.LoadUint64(&add) != 2 {
			t.Errorf("expected 2 events of type Create but %v occurred", add)
		}
		// secret sync triggers a dummy event
		if atomic.LoadUint64(&upd) != 1 {
			t.Errorf("expected 1 events of type Update but %v occurred", upd)
		}

		err = clientSet.CoreV1().Secrets(ns).Delete(secretName, &metav1.DeleteOptions{})
		if err != nil {
			t.Errorf("error deleting secret: %v", err)
		}

		time.Sleep(1 * time.Second)

		if atomic.LoadUint64(&del) != 1 {
			t.Errorf("expected 1 events of type Delete but %v occurred", del)
		}

	})

	t.Run("should create an ingress with a secret which does not exist", func(t *testing.T) {
		ns := createNamespace(clientSet, t)
		defer deleteNamespace(ns, clientSet, t)
		cm := createConfigMap(clientSet, ns, t)
		defer deleteConfigMap(cm, ns, clientSet, t)

		stopCh := make(chan struct{})
		updateCh := channels.NewRingChannel(1024)

		var add uint64
		var upd uint64
		var del uint64

		go func(ch *channels.RingChannel) {
			for {
				evt, ok := <-ch.Out()
				if !ok {
					return
				}

				e := evt.(Event)
				if e.Obj == nil {
					continue
				}
				switch e.Type {
				case CreateEvent:
					atomic.AddUint64(&add, 1)
				case UpdateEvent:
					atomic.AddUint64(&upd, 1)
				case DeleteEvent:
					atomic.AddUint64(&del, 1)
				}
			}
		}(updateCh)

		fs := newFS(t)
		storer := New(true,
			ns,
			fmt.Sprintf("%v/config", ns),
			fmt.Sprintf("%v/tcp", ns),
			fmt.Sprintf("%v/udp", ns),
			"",
			10*time.Minute,
			clientSet,
			fs,
			updateCh,
			false,
			pod,
			false)

		storer.Run(stopCh)

		name := "ingress-with-secret"
		secretHosts := []string{name}

		ing := ensureIngress(&extensions.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: ns,
				SelfLink:  fmt.Sprintf("/apis/extensions/v1beta1/namespaces/%s/ingresses/%s", ns, name),
			},
			Spec: extensions.IngressSpec{
				TLS: []extensions.IngressTLS{
					{
						Hosts:      secretHosts,
						SecretName: name,
					},
				},
				Rules: []extensions.IngressRule{
					{
						Host: name,
						IngressRuleValue: extensions.IngressRuleValue{
							HTTP: &extensions.HTTPIngressRuleValue{
								Paths: []extensions.HTTPIngressPath{
									{
										Path: "/",
										Backend: extensions.IngressBackend{
											ServiceName: "http-svc",
											ServicePort: intstr.FromInt(80),
										},
									},
								},
							},
						},
					},
				},
			},
		}, clientSet, t)
		defer deleteIngress(ing, clientSet, t)

		err := framework.WaitForIngressInNamespace(clientSet, ns, name)
		if err != nil {
			t.Errorf("error waiting for ingress: %v", err)
		}

		// take into account delay caused by:
		//  * ingress annotations extraction
		//  * secretIngressMap update
		//  * secrets sync
		time.Sleep(3 * time.Second)

		if atomic.LoadUint64(&add) != 1 {
			t.Errorf("expected 1 events of type Create but %v occurred", add)
		}
		if atomic.LoadUint64(&upd) != 0 {
			t.Errorf("expected 0 events of type Update but %v occurred", upd)
		}
		if atomic.LoadUint64(&del) != 0 {
			t.Errorf("expected 0 events of type Delete but %v occurred", del)
		}

		_, err = framework.CreateIngressTLSSecret(clientSet, secretHosts, name, ns)
		if err != nil {
			t.Errorf("error creating secret: %v", err)
		}

		t.Run("should exists a secret in the local store and filesystem", func(t *testing.T) {
			err := framework.WaitForSecretInNamespace(clientSet, ns, name)
			if err != nil {
				t.Errorf("error waiting for secret: %v", err)
			}

			time.Sleep(5 * time.Second)

			pemFile := fmt.Sprintf("%v/%v-%v.pem", file.DefaultSSLDirectory, ns, name)
			err = framework.WaitForFileInFS(pemFile, fs)
			if err != nil {
				t.Errorf("error waiting for file to exist on the file system: %v", err)
			}

			secretName := fmt.Sprintf("%v/%v", ns, name)
			sslCert, err := storer.GetLocalSSLCert(secretName)
			if err != nil {
				t.Errorf("error reading local secret %v: %v", secretName, err)
			}

			if sslCert == nil {
				t.Errorf("expected a secret but none returned")
			}

			pemSHA := file.SHA1(pemFile)
			if sslCert.PemSHA != pemSHA {
				t.Errorf("SHA of secret on disk differs from local secret store (%v != %v)", pemSHA, sslCert.PemSHA)
			}
		})
	})

	// test add ingress with secret it doesn't exists and then add secret
	// check secret is generated on fs
	// check ocsp
	// check invalid secret (missing crt)
	// check invalid secret (missing key)
	// check invalid secret (missing ca)
}

func createNamespace(clientSet kubernetes.Interface, t *testing.T) string {
	t.Helper()
	t.Log("Creating temporal namespace")

	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "store-test",
		},
	}

	ns, err := clientSet.CoreV1().Namespaces().Create(namespace)
	if err != nil {
		t.Errorf("error creating the namespace: %v", err)
	}
	t.Logf("Temporal namespace %v created", ns)

	return ns.Name
}

func deleteNamespace(ns string, clientSet kubernetes.Interface, t *testing.T) {
	t.Helper()
	t.Logf("Deleting temporal namespace %v", ns)

	err := clientSet.CoreV1().Namespaces().Delete(ns, &metav1.DeleteOptions{})
	if err != nil {
		t.Errorf("error deleting the namespace: %v", err)
	}
	t.Logf("Temporal namespace %v deleted", ns)
}

func createConfigMap(clientSet kubernetes.Interface, ns string, t *testing.T) string {
	t.Helper()
	t.Log("Creating temporal config map")

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:     "config",
			SelfLink: fmt.Sprintf("/api/v1/namespaces/%s/configmaps/config", ns),
		},
	}

	cm, err := clientSet.CoreV1().ConfigMaps(ns).Create(configMap)
	if err != nil {
		t.Errorf("error creating the configuration map: %v", err)
	}
	t.Logf("Temporal configmap %v created", cm)

	return cm.Name
}

func deleteConfigMap(cm, ns string, clientSet kubernetes.Interface, t *testing.T) {
	t.Helper()
	t.Logf("Deleting temporal configmap %v", cm)

	err := clientSet.CoreV1().ConfigMaps(ns).Delete(cm, &metav1.DeleteOptions{})
	if err != nil {
		t.Errorf("error deleting the configmap: %v", err)
	}
	t.Logf("Temporal configmap %v deleted", cm)
}

func ensureIngress(ingress *extensions.Ingress, clientSet kubernetes.Interface, t *testing.T) *extensions.Ingress {
	t.Helper()
	ing, err := clientSet.Extensions().Ingresses(ingress.Namespace).Update(ingress)

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			t.Logf("Ingress %v not found, creating", ingress)

			ing, err = clientSet.Extensions().Ingresses(ingress.Namespace).Create(ingress)
			if err != nil {
				t.Fatalf("error creating ingress %+v: %v", ingress, err)
			}

			t.Logf("Ingress %+v created", ingress)
			return ing
		}

		t.Fatalf("error updating ingress %+v: %v", ingress, err)
	}

	t.Logf("Ingress %+v updated", ingress)

	return ing
}

func deleteIngress(ingress *extensions.Ingress, clientSet kubernetes.Interface, t *testing.T) {
	t.Helper()
	err := clientSet.Extensions().Ingresses(ingress.Namespace).Delete(ingress.Name, &metav1.DeleteOptions{})

	if err != nil {
		t.Errorf("failed to delete ingress %+v: %v", ingress, err)
	}

	t.Logf("Ingress %+v deleted", ingress)
}

func newFS(t *testing.T) file.Filesystem {
	fs, err := file.NewFakeFS()
	if err != nil {
		t.Fatalf("error creating filesystem: %v", err)
	}
	return fs
}

// newStore creates a new mock object store for tests which do not require the
// use of Informers.
func newStore(t *testing.T) *k8sStore {
	fs, err := file.NewFakeFS()
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	pod := &k8s.PodInfo{
		Name:      "ingress-1",
		Namespace: v1.NamespaceDefault,
		Labels: map[string]string{
			"pod-template-hash": "1234",
		},
	}

	return &k8sStore{
		listers: &Lister{
			// add more listers if needed
			Ingress:               IngressLister{cache.NewStore(cache.MetaNamespaceKeyFunc)},
			IngressWithAnnotation: IngressWithAnnotationsLister{cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)},
			Pod:                   PodLister{cache.NewStore(cache.MetaNamespaceKeyFunc)},
		},
		sslStore:         NewSSLCertTracker(),
		filesystem:       fs,
		updateCh:         channels.NewRingChannel(10),
		syncSecretMu:     new(sync.Mutex),
		backendConfigMu:  new(sync.RWMutex),
		secretIngressMap: NewObjectRefMap(),
		pod:              pod,
	}
}

func TestUpdateSecretIngressMap(t *testing.T) {
	s := newStore(t)

	ingTpl := &extensions.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "testns",
		},
	}
	s.listers.Ingress.Add(ingTpl)

	t.Run("with TLS secret", func(t *testing.T) {
		ing := ingTpl.DeepCopy()
		ing.Spec = extensions.IngressSpec{
			TLS: []extensions.IngressTLS{{SecretName: "tls"}},
		}
		s.listers.Ingress.Update(ing)
		s.updateSecretIngressMap(ing)

		if l := s.secretIngressMap.Len(); !(l == 1 && s.secretIngressMap.Has("testns/tls")) {
			t.Errorf("Expected \"testns/tls\" to be the only referenced Secret (got %d)", l)
		}
	})

	t.Run("with annotation in simple name format", func(t *testing.T) {
		ing := ingTpl.DeepCopy()
		ing.ObjectMeta.SetAnnotations(map[string]string{
			parser.GetAnnotationWithPrefix("auth-secret"): "auth",
		})
		s.listers.Ingress.Update(ing)
		s.updateSecretIngressMap(ing)

		if l := s.secretIngressMap.Len(); !(l == 1 && s.secretIngressMap.Has("testns/auth")) {
			t.Errorf("Expected \"testns/auth\" to be the only referenced Secret (got %d)", l)
		}
	})

	t.Run("with annotation in namespace/name format", func(t *testing.T) {
		ing := ingTpl.DeepCopy()
		ing.ObjectMeta.SetAnnotations(map[string]string{
			parser.GetAnnotationWithPrefix("auth-secret"): "otherns/auth",
		})
		s.listers.Ingress.Update(ing)
		s.updateSecretIngressMap(ing)

		if l := s.secretIngressMap.Len(); !(l == 1 && s.secretIngressMap.Has("otherns/auth")) {
			t.Errorf("Expected \"otherns/auth\" to be the only referenced Secret (got %d)", l)
		}
	})

	t.Run("with annotation in invalid format", func(t *testing.T) {
		ing := ingTpl.DeepCopy()
		ing.ObjectMeta.SetAnnotations(map[string]string{
			parser.GetAnnotationWithPrefix("auth-secret"): "ns/name/garbage",
		})
		s.listers.Ingress.Update(ing)
		s.updateSecretIngressMap(ing)

		if l := s.secretIngressMap.Len(); l != 0 {
			t.Errorf("Expected 0 referenced Secret (got %d)", l)
		}
	})
}

func TestListIngresses(t *testing.T) {
	s := newStore(t)

	ingressToIgnore := &ingress.Ingress{
		Ingress: extensions.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-2",
				Namespace: "testns",
				Annotations: map[string]string{
					"kubernetes.io/ingress.class": "something",
				},
				CreationTimestamp: metav1.NewTime(time.Now()),
			},
			Spec: extensions.IngressSpec{
				Backend: &extensions.IngressBackend{
					ServiceName: "demo",
					ServicePort: intstr.FromInt(80),
				},
			},
		},
	}
	s.listers.IngressWithAnnotation.Add(ingressToIgnore)

	ingressWithoutPath := &ingress.Ingress{
		Ingress: extensions.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test-3",
				Namespace:         "testns",
				CreationTimestamp: metav1.NewTime(time.Now()),
			},
			Spec: extensions.IngressSpec{
				Rules: []extensions.IngressRule{
					{
						Host: "foo.bar",
						IngressRuleValue: extensions.IngressRuleValue{
							HTTP: &extensions.HTTPIngressRuleValue{
								Paths: []extensions.HTTPIngressPath{
									{
										Backend: extensions.IngressBackend{
											ServiceName: "demo",
											ServicePort: intstr.FromInt(80),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	s.listers.IngressWithAnnotation.Add(ingressWithoutPath)

	ingressWithNginxClass := &ingress.Ingress{
		Ingress: extensions.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-4",
				Namespace: "testns",
				Annotations: map[string]string{
					"kubernetes.io/ingress.class": "nginx",
				},
				CreationTimestamp: metav1.NewTime(time.Now()),
			},
			Spec: extensions.IngressSpec{
				Rules: []extensions.IngressRule{
					{
						Host: "foo.bar",
						IngressRuleValue: extensions.IngressRuleValue{
							HTTP: &extensions.HTTPIngressRuleValue{
								Paths: []extensions.HTTPIngressPath{
									{
										Path: "/demo",
										Backend: extensions.IngressBackend{
											ServiceName: "demo",
											ServicePort: intstr.FromInt(80),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	s.listers.IngressWithAnnotation.Add(ingressWithNginxClass)

	ingresses := s.ListIngresses()

	if s := len(ingresses); s != 3 {
		t.Errorf("Expected 3 Ingresses but got %v", s)
	}

	if ingresses[0].Name != "test-2" {
		t.Errorf("Expected Ingress test-2 but got %v", ingresses[0].Name)
	}

	if ingresses[1].Name != "test-3" {
		t.Errorf("Expected Ingress test-3 but got %v", ingresses[1].Name)
	}

	if ingresses[2].Name != "test-4" {
		t.Errorf("Expected Ingress test-4 but got %v", ingresses[2].Name)
	}
}

func TestWriteSSLSessionTicketKey(t *testing.T) {
	tests := []string{
		"9DyULjtYWz520d1rnTLbc4BOmN2nLAVfd3MES/P3IxWuwXkz9Fby0lnOZZUdNEMV",
		"9SvN1C9AB5DvNde5fMKoJwAwICpqdjiMyxR+cv6NpAWv22rFd3gKt4wMyGxCm7l9Wh6BQPG0+csyBZSHHr2NOWj52Wx8xCegXf4NsSMBUqA=",
	}

	for _, test := range tests {
		s := newStore(t)

		cmap := &v1.ConfigMap{
			Data: map[string]string{
				"ssl-session-ticket-key": test,
			},
		}

		f, err := ioutil.TempFile("", "ssl-session-ticket-test")
		if err != nil {
			t.Fatal(err)
		}

		s.writeSSLSessionTicketKey(cmap, f.Name())

		content, err := ioutil.ReadFile(f.Name())
		if err != nil {
			t.Fatal(err)
		}
		encodedContent := base64.StdEncoding.EncodeToString(content)

		f.Close()

		if test != encodedContent {
			t.Fatalf("expected %v but returned %s", test, encodedContent)
		}
	}
}

func TestGetRunningControllerPodsCount(t *testing.T) {
	os.Setenv("POD_NAMESPACE", "testns")
	os.Setenv("POD_NAME", "ingress-1")

	s := newStore(t)
	s.pod = &k8s.PodInfo{
		Name:      "ingress-1",
		Namespace: "testns",
		Labels: map[string]string{
			"pod-template-hash": "1234",
		},
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ingress-1",
			Namespace: "testns",
			Labels: map[string]string{
				"pod-template-hash": "1234",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	s.listers.Pod.Add(pod)

	pod = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ingress-2",
			Namespace: "testns",
			Labels: map[string]string{
				"pod-template-hash": "1234",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	s.listers.Pod.Add(pod)

	pod = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ingress-3",
			Namespace: "testns",
			Labels: map[string]string{
				"pod-template-hash": "1234",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodFailed,
		},
	}
	s.listers.Pod.Add(pod)

	podsCount := s.GetRunningControllerPodsCount()
	if podsCount != 2 {
		t.Errorf("Expected 1 controller Pods but got %v", s)
	}
}
