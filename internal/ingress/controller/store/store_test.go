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
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/eapache/channels"
	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	networking "k8s.io/api/networking/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations/class"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

func TestStore(t *testing.T) {
	k8s.IsNetworkingIngressAvailable = true

	pod := &k8s.PodInfo{
		Name:      "testpod",
		Namespace: v1.NamespaceDefault,
		Labels: map[string]string{
			"pod-template-hash": "1234",
		},
	}

	//TODO: move env definition to docker image?
	os.Setenv("KUBEBUILDER_ASSETS", "/usr/local/bin")

	te := &envtest.Environment{}
	cfg, err := te.Start()
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	defer te.Stop()

	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	t.Run("should return an error searching for non existing objects", func(t *testing.T) {
		ns := createNamespace(clientSet, t)
		defer deleteNamespace(ns, clientSet, t)
		createConfigMap(clientSet, ns, t)

		stopCh := make(chan struct{})
		updateCh := channels.NewRingChannel(1024)

		go func(ch *channels.RingChannel) {
			for {
				<-ch.Out()
			}
		}(updateCh)

		storer := New(
			ns,
			fmt.Sprintf("%v/config", ns),
			fmt.Sprintf("%v/tcp", ns),
			fmt.Sprintf("%v/udp", ns),
			"",
			10*time.Minute,
			clientSet,
			updateCh,
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
		createConfigMap(clientSet, ns, t)

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
				if _, ok := e.Obj.(*networking.Ingress); !ok {
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

		storer := New(
			ns,
			fmt.Sprintf("%v/config", ns),
			fmt.Sprintf("%v/tcp", ns),
			fmt.Sprintf("%v/udp", ns),
			"",
			10*time.Minute,
			clientSet,
			updateCh,
			pod,
			false)

		storer.Run(stopCh)

		ing := ensureIngress(&networking.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dummy",
				Namespace: ns,
			},
			Spec: networking.IngressSpec{
				Rules: []networking.IngressRule{
					{
						Host: "dummy",
						IngressRuleValue: networking.IngressRuleValue{
							HTTP: &networking.HTTPIngressRuleValue{
								Paths: []networking.HTTPIngressPath{
									{
										Path: "/",
										Backend: networking.IngressBackend{
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
		invalidIngress := ensureIngress(&networking.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "custom-class",
				Namespace: ns,
				Annotations: map[string]string{
					class.IngressKey: "something",
				},
			},
			Spec: networking.IngressSpec{
				Rules: []networking.IngressRule{
					{
						Host: "dummy",
						IngressRuleValue: networking.IngressRuleValue{
							HTTP: &networking.HTTPIngressRuleValue{
								Paths: []networking.HTTPIngressPath{
									{
										Path: "/",
										Backend: networking.IngressBackend{
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

		err = clientSet.NetworkingV1beta1().Ingresses(ni.Namespace).Delete(context.TODO(), ni.Name, metav1.DeleteOptions{})
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
		createConfigMap(clientSet, ns, t)

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
				if _, ok := e.Obj.(*networking.Ingress); !ok {
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

		storer := New(
			ns,
			fmt.Sprintf("%v/config", ns),
			fmt.Sprintf("%v/tcp", ns),
			fmt.Sprintf("%v/udp", ns),
			"",
			10*time.Minute,
			clientSet,
			updateCh,
			pod,
			false)

		storer.Run(stopCh)

		// create an invalid ingress (different class)
		invalidIngress := ensureIngress(&networking.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "custom-class",
				Namespace: ns,
				Annotations: map[string]string{
					class.IngressKey: "something",
				},
			},
			Spec: networking.IngressSpec{
				Rules: []networking.IngressRule{
					{
						Host: "dummy",
						IngressRuleValue: networking.IngressRuleValue{
							HTTP: &networking.HTTPIngressRuleValue{
								Paths: []networking.HTTPIngressPath{
									{
										Path: "/",
										Backend: networking.IngressBackend{
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
		createConfigMap(clientSet, ns, t)

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

		storer := New(
			ns,
			fmt.Sprintf("%v/config", ns),
			fmt.Sprintf("%v/tcp", ns),
			fmt.Sprintf("%v/udp", ns),
			"",
			10*time.Minute,
			clientSet,
			updateCh,
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

		err = clientSet.CoreV1().Secrets(ns).Delete(context.TODO(), secretName, metav1.DeleteOptions{})
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
		createConfigMap(clientSet, ns, t)

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

		storer := New(
			ns,
			fmt.Sprintf("%v/config", ns),
			fmt.Sprintf("%v/tcp", ns),
			fmt.Sprintf("%v/udp", ns),
			"",
			10*time.Minute,
			clientSet,
			updateCh,
			pod,
			false)

		storer.Run(stopCh)

		ingressName := "ingress-with-secret"
		secretName := "referenced"

		ing := ensureIngress(&networking.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ingressName,
				Namespace: ns,
			},
			Spec: networking.IngressSpec{
				TLS: []networking.IngressTLS{
					{
						SecretName: secretName,
					},
				},
				Backend: &networking.IngressBackend{
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

		err = clientSet.CoreV1().Secrets(ns).Delete(context.TODO(), secretName, metav1.DeleteOptions{})
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
		createConfigMap(clientSet, ns, t)

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

		storer := New(
			ns,
			fmt.Sprintf("%v/config", ns),
			fmt.Sprintf("%v/tcp", ns),
			fmt.Sprintf("%v/udp", ns),
			"",
			10*time.Minute,
			clientSet,
			updateCh,
			pod,
			false)

		storer.Run(stopCh)

		name := "ingress-with-secret"
		secretHosts := []string{name}

		ing := ensureIngress(&networking.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: ns,
			},
			Spec: networking.IngressSpec{
				TLS: []networking.IngressTLS{
					{
						Hosts:      secretHosts,
						SecretName: name,
					},
				},
				Rules: []networking.IngressRule{
					{
						Host: name,
						IngressRuleValue: networking.IngressRuleValue{
							HTTP: &networking.HTTPIngressRuleValue{
								Paths: []networking.HTTPIngressPath{
									{
										Path: "/",
										Backend: networking.IngressBackend{
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

	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("store-test-%v", time.Now().Unix()),
		},
	}

	ns, err := clientSet.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	if err != nil {
		t.Errorf("error creating the namespace: %v", err)
	}

	return ns.Name
}

func deleteNamespace(ns string, clientSet kubernetes.Interface, t *testing.T) {
	t.Helper()

	err := clientSet.CoreV1().Namespaces().Delete(context.TODO(), ns, metav1.DeleteOptions{})
	if err != nil {
		t.Errorf("error deleting the namespace: %v", err)
	}
}

func createConfigMap(clientSet kubernetes.Interface, ns string, t *testing.T) string {
	t.Helper()

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "config",
		},
	}

	cm, err := clientSet.CoreV1().ConfigMaps(ns).Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		t.Errorf("error creating the configuration map: %v", err)
	}

	return cm.Name
}

func ensureIngress(ingress *networking.Ingress, clientSet kubernetes.Interface, t *testing.T) *networking.Ingress {
	t.Helper()
	ing, err := clientSet.NetworkingV1beta1().Ingresses(ingress.Namespace).Update(context.TODO(), ingress, metav1.UpdateOptions{})

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			t.Logf("Ingress %v not found, creating", ingress)

			ing, err = clientSet.NetworkingV1beta1().Ingresses(ingress.Namespace).Create(context.TODO(), ingress, metav1.CreateOptions{})
			if err != nil {
				t.Fatalf("error creating ingress %+v: %v", ingress, err)
			}

			t.Logf("Ingress %+v created", ingress)
			return ing
		}

		t.Fatalf("error updating ingress %+v: %v", ingress, err)
	}

	return ing
}

func deleteIngress(ingress *networking.Ingress, clientSet kubernetes.Interface, t *testing.T) {
	t.Helper()
	err := clientSet.NetworkingV1beta1().Ingresses(ingress.Namespace).Delete(context.TODO(), ingress.Name, metav1.DeleteOptions{})

	if err != nil {
		t.Errorf("failed to delete ingress %+v: %v", ingress, err)
	}

	t.Logf("Ingress %+v deleted", ingress)
}

// newStore creates a new mock object store for tests which do not require the
// use of Informers.
func newStore(t *testing.T) *k8sStore {
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
		updateCh:         channels.NewRingChannel(10),
		syncSecretMu:     new(sync.Mutex),
		backendConfigMu:  new(sync.RWMutex),
		secretIngressMap: NewObjectRefMap(),
		pod:              pod,
	}
}

func TestUpdateSecretIngressMap(t *testing.T) {
	s := newStore(t)

	ingTpl := &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "testns",
		},
	}
	s.listers.Ingress.Add(ingTpl)

	t.Run("with TLS secret", func(t *testing.T) {
		ing := ingTpl.DeepCopy()
		ing.Spec = networking.IngressSpec{
			TLS: []networking.IngressTLS{{SecretName: "tls"}},
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
		Ingress: networking.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-2",
				Namespace: "testns",
				Annotations: map[string]string{
					class.IngressKey: "something",
				},
				CreationTimestamp: metav1.NewTime(time.Now()),
			},
			Spec: networking.IngressSpec{
				Backend: &networking.IngressBackend{
					ServiceName: "demo",
					ServicePort: intstr.FromInt(80),
				},
			},
		},
	}
	s.listers.IngressWithAnnotation.Add(ingressToIgnore)

	ingressWithoutPath := &ingress.Ingress{
		Ingress: networking.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test-3",
				Namespace:         "testns",
				CreationTimestamp: metav1.NewTime(time.Now()),
			},
			Spec: networking.IngressSpec{
				Rules: []networking.IngressRule{
					{
						Host: "foo.bar",
						IngressRuleValue: networking.IngressRuleValue{
							HTTP: &networking.HTTPIngressRuleValue{
								Paths: []networking.HTTPIngressPath{
									{
										Backend: networking.IngressBackend{
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
		Ingress: networking.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-4",
				Namespace: "testns",
				Annotations: map[string]string{
					class.IngressKey: "nginx",
				},
				CreationTimestamp: metav1.NewTime(time.Now()),
			},
			Spec: networking.IngressSpec{
				Rules: []networking.IngressRule{
					{
						Host: "foo.bar",
						IngressRuleValue: networking.IngressRuleValue{
							HTTP: &networking.HTTPIngressRuleValue{
								Paths: []networking.HTTPIngressPath{
									{
										Path: "/demo",
										Backend: networking.IngressBackend{
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

	ingresses := s.ListIngresses(nil)

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

func TestIngressConversion(t *testing.T) {
	ing := &extensions.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "old-ingress",
			Namespace:         "demo",
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
	}

	new, err := fromExtensions(ing)
	if err != nil {
		t.Fatalf("unexpected error converting ingress: %v", err)
	}

	m1, err := new.Marshal()
	if err != nil {
		t.Fatalf("unexpected error marshalling Ingress: %v", err)
	}

	m2, err := ing.Marshal()
	if err != nil {
		t.Fatalf("unexpected error marshalling Ingress: %v", err)
	}

	if !bytes.Equal(m1, m2) {
		t.Fatalf("Expected marshalling of types should be equal")
	}
}
