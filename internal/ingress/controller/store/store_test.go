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
	"k8s.io/api/extensions/v1beta1"
	extensions "k8s.io/api/extensions/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"k8s.io/ingress-nginx/internal/file"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

func TestStore(t *testing.T) {
	// TODO: find a way to avoid the need to use a real api server
	home := os.Getenv("HOME")
	kubeConfigFile := fmt.Sprintf("%v/.kube/config", home)
	kubeContext := ""

	kubeConfig, err := framework.LoadConfig(kubeConfigFile, kubeContext)
	if err != nil {
		t.Errorf("unexpected error loading kubeconfig file: %v", err)
	}

	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		t.Errorf("unexpected error creating ingress client: %v", err)
	}

	t.Run("should return an error searching for non existing objects", func(t *testing.T) {
		ns := createNamespace(clientSet, t)
		defer deleteNamespace(ns, clientSet, t)

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
			updateCh)

		storer.Run(stopCh)

		key := fmt.Sprintf("%v/anything", ns)
		ing, err := storer.GetIngress(key)
		if err == nil {
			t.Errorf("expected an error but none returned")
		}
		if ing != nil {
			t.Errorf("expected an Ingres but none returned")
		}

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

		updateCh.Close()
		close(stopCh)
	})

	t.Run("should return one event for add, update and delete of ingress", func(t *testing.T) {
		ns := createNamespace(clientSet, t)
		defer deleteNamespace(ns, clientSet, t)

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
					t.Errorf("expected an Ingress type but %T returned", e.Obj)
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
			updateCh)

		storer.Run(stopCh)

		ing, err := ensureIngress(&v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dummy",
				Namespace: ns,
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "dummy",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/",
										Backend: v1beta1.IngressBackend{
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
		}, clientSet)
		if err != nil {
			t.Errorf("unexpected error creating ingress: %v", err)
		}

		// create an invalid ingress (different class)
		_, err = ensureIngress(&v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "custom-class",
				Namespace: ns,
				Annotations: map[string]string{
					"kubernetes.io/ingress.class": "something",
				},
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "dummy",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/",
										Backend: v1beta1.IngressBackend{
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
		}, clientSet)
		if err != nil {
			t.Errorf("unexpected error creating ingress: %v", err)
		}

		ni := ing.DeepCopy()
		ni.Spec.Rules[0].Host = "update-dummy"
		_, err = ensureIngress(ni, clientSet)
		if err != nil {
			t.Errorf("unexpected error creating ingress: %v", err)
		}

		err = clientSet.ExtensionsV1beta1().
			Ingresses(ni.Namespace).
			Delete(ni.Name, &metav1.DeleteOptions{})
		if err != nil {
			t.Errorf("unexpected error creating ingress: %v", err)
		}

		framework.WaitForNoIngressInNamespace(clientSet, ni.Namespace, ni.Name)

		if atomic.LoadUint64(&add) != 1 {
			t.Errorf("expected 1 event of type Create but %v occurred", add)
		}
		if atomic.LoadUint64(&upd) != 1 {
			t.Errorf("expected 1 event of type Update but %v occurred", upd)
		}
		if atomic.LoadUint64(&del) != 1 {
			t.Errorf("expected 1 event of type Delete but %v occurred", del)
		}

		updateCh.Close()
		close(stopCh)
	})

	t.Run("should not receive events from secret not referenced from ingress", func(t *testing.T) {
		ns := createNamespace(clientSet, t)
		defer deleteNamespace(ns, clientSet, t)

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
			updateCh)

		storer.Run(stopCh)

		secretName := "not-referenced"
		_, _, _, err = framework.CreateIngressTLSSecret(clientSet, []string{"foo"}, secretName, ns)
		if err != nil {
			t.Errorf("unexpected error creating secret: %v", err)
		}

		err = framework.WaitForSecretInNamespace(clientSet, ns, secretName)
		if err != nil {
			t.Errorf("unexpected error waiting for secret: %v", err)
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
			t.Errorf("unexpected error deleting secret: %v", err)
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

		updateCh.Close()
		close(stopCh)
	})

	t.Run("should receive events from secret referenced from ingress", func(t *testing.T) {
		ns := createNamespace(clientSet, t)
		defer deleteNamespace(ns, clientSet, t)

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
			updateCh)

		storer.Run(stopCh)

		ingressName := "ingress-with-secret"
		secretName := "referenced"

		_, err := ensureIngress(&v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ingressName,
				Namespace: ns,
			},
			Spec: v1beta1.IngressSpec{
				TLS: []v1beta1.IngressTLS{
					{
						SecretName: secretName,
					},
				},
				Backend: &v1beta1.IngressBackend{
					ServiceName: "http-svc",
					ServicePort: intstr.FromInt(80),
				},
			},
		}, clientSet)
		if err != nil {
			t.Errorf("unexpected error creating ingress: %v", err)
		}

		err = framework.WaitForIngressInNamespace(clientSet, ns, ingressName)
		if err != nil {
			t.Errorf("unexpected error waiting for secret: %v", err)
		}

		_, _, _, err = framework.CreateIngressTLSSecret(clientSet, []string{"foo"}, secretName, ns)
		if err != nil {
			t.Errorf("unexpected error creating secret: %v", err)
		}

		err = framework.WaitForSecretInNamespace(clientSet, ns, secretName)
		if err != nil {
			t.Errorf("unexpected error waiting for secret: %v", err)
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
			t.Errorf("unexpected error deleting secret: %v", err)
		}

		time.Sleep(1 * time.Second)

		if atomic.LoadUint64(&del) != 1 {
			t.Errorf("expected 1 events of type Delete but %v occurred", del)
		}

		updateCh.Close()
		close(stopCh)
	})

	t.Run("should create an ingress with a secret which does not exist", func(t *testing.T) {
		ns := createNamespace(clientSet, t)
		defer deleteNamespace(ns, clientSet, t)

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
			updateCh)

		storer.Run(stopCh)

		name := "ingress-with-secret"
		secretHosts := []string{name}

		_, err := ensureIngress(&v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: ns,
			},
			Spec: v1beta1.IngressSpec{
				TLS: []v1beta1.IngressTLS{
					{
						Hosts:      secretHosts,
						SecretName: name,
					},
				},
				Rules: []v1beta1.IngressRule{
					{
						Host: name,
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/",
										Backend: v1beta1.IngressBackend{
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
		}, clientSet)
		if err != nil {
			t.Errorf("unexpected error creating ingress: %v", err)
		}

		err = framework.WaitForIngressInNamespace(clientSet, ns, name)
		if err != nil {
			t.Errorf("unexpected error waiting for ingress: %v", err)
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

		_, _, _, err = framework.CreateIngressTLSSecret(clientSet, secretHosts, name, ns)
		if err != nil {
			t.Errorf("unexpected error creating secret: %v", err)
		}

		t.Run("should exists a secret in the local store and filesystem", func(t *testing.T) {
			err := framework.WaitForSecretInNamespace(clientSet, ns, name)
			if err != nil {
				t.Errorf("unexpected error waiting for secret: %v", err)
			}

			time.Sleep(5 * time.Second)

			pemFile := fmt.Sprintf("%v/%v-%v.pem", file.DefaultSSLDirectory, ns, name)
			err = framework.WaitForFileInFS(pemFile, fs)
			if err != nil {
				t.Errorf("unexpected error waiting for file to exist on the file system: %v", err)
			}

			secretName := fmt.Sprintf("%v/%v", ns, name)
			sslCert, err := storer.GetLocalSSLCert(secretName)
			if err != nil {
				t.Errorf("unexpected error reading local secret %v: %v", secretName, err)
			}

			if sslCert == nil {
				t.Errorf("expected a secret but none returned")
			}

			pemSHA := file.SHA1(pemFile)
			if sslCert.PemSHA != pemSHA {
				t.Errorf("SHA of secret on disk differs from local secret store (%v != %v)", pemSHA, sslCert.PemSHA)
			}
		})

		updateCh.Close()
		close(stopCh)
	})

	// test add ingress with secret it doesn't exists and then add secret
	// check secret is generated on fs
	// check ocsp
	// check invalid secret (missing crt)
	// check invalid secret (missing key)
	// check invalid secret (missing ca)
}

func createNamespace(clientSet *kubernetes.Clientset, t *testing.T) string {
	t.Log("creating temporal namespace")
	ns, err := framework.CreateKubeNamespace("store-test", clientSet)
	if err != nil {
		t.Errorf("unexpected error creating ingress client: %v", err)
	}
	t.Logf("temporal namespace %v created", ns)

	return ns
}

func deleteNamespace(ns string, clientSet *kubernetes.Clientset, t *testing.T) {
	t.Logf("deleting temporal namespace %v created", ns)
	err := framework.DeleteKubeNamespace(clientSet, ns)
	if err != nil {
		t.Errorf("unexpected error creating ingress client: %v", err)
	}
	t.Logf("temporal namespace %v deleted", ns)
}

func ensureIngress(ingress *extensions.Ingress, clientSet *kubernetes.Clientset) (*extensions.Ingress, error) {
	s, err := clientSet.ExtensionsV1beta1().Ingresses(ingress.Namespace).Update(ingress)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return clientSet.ExtensionsV1beta1().Ingresses(ingress.Namespace).Create(ingress)
		}
		return nil, err
	}
	return s, nil
}

func newFS(t *testing.T) file.Filesystem {
	fs, err := file.NewFakeFS()
	if err != nil {
		t.Fatalf("unexpected error creating filesystem: %v", err)
	}
	return fs
}

// newStore creates a new mock object store for tests which do not require the
// use of Informers.
func newStore(t *testing.T) *k8sStore {
	fs, err := file.NewFakeFS()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return &k8sStore{
		listers: &Lister{
			// add more listers if needed
			Ingress: IngressLister{cache.NewStore(cache.MetaNamespaceKeyFunc)},
		},
		sslStore:         NewSSLCertTracker(),
		filesystem:       fs,
		updateCh:         channels.NewRingChannel(10),
		mu:               new(sync.Mutex),
		secretIngressMap: NewObjectRefMap(),
	}
}

func TestUpdateSecretIngressMap(t *testing.T) {
	s := newStore(t)

	ingTpl := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "testns",
		},
	}
	s.listers.Ingress.Add(ingTpl)

	t.Run("with TLS secret", func(t *testing.T) {
		ing := ingTpl.DeepCopy()
		ing.Spec = v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{{SecretName: "tls"}},
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
