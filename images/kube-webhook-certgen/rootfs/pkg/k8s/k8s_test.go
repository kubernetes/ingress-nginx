package k8s

import (
	"bytes"
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	aggregatorfake "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/fake"
)

const (
	testWebhookName    = "c7c95710-d8c3-4cc3-a2a8-8d2b46909c76"
	testSecretName     = "15906410-af2a-4f9b-8a2d-c08ffdd5e129"
	testAPIServiceName = "37f6a2d1-b401-4275-833b-9ff5004f0301"
	testNamespace      = "7cad5f92-c0d5-4bc9-87a3-6f44d5a5619d"
)

var (
	fail   = admissionv1.Fail
	ignore = admissionv1.Ignore
)

func genSecretData() (ca, cert, key []byte) {
	ca = make([]byte, 4)
	cert = make([]byte, 4)
	key = make([]byte, 4)
	rand.Read(cert)
	rand.Read(key)
	return
}

func newTestSimpleK8s(objects ...runtime.Object) *k8s {
	return &k8s{
		clientset:           fake.NewSimpleClientset(objects...),
		aggregatorClientset: aggregatorfake.NewSimpleClientset(),
	}
}

func TestGetCaFromCertificate(t *testing.T) {
	ca, cert, key := genSecretData()

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testSecretName,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{"ca": ca, "cert": cert, "key": key},
	}

	k := newTestSimpleK8s(secret)

	retrievedCa := k.GetCaFromSecret(contextWithDeadline(t), testSecretName, testNamespace)
	if !bytes.Equal(retrievedCa, ca) {
		t.Error("Was not able to retrieve CA information that was saved")
	}
}

func TestSaveCertsToSecret(t *testing.T) {
	k := newTestSimpleK8s()

	ca, cert, key := genSecretData()

	ctx := contextWithDeadline(t)

	k.SaveCertsToSecret(ctx, testSecretName, testNamespace, "cert", "key", ca, cert, key)

	secret, _ := k.clientset.CoreV1().Secrets(testNamespace).Get(ctx, testSecretName, metav1.GetOptions{})

	if !bytes.Equal(secret.Data["cert"], cert) {
		t.Error("'cert' saved data does not match retrieved")
	}

	if !bytes.Equal(secret.Data["key"], key) {
		t.Error("'key' saved data does not match retrieved")
	}
}

func TestSaveThenLoadSecret(t *testing.T) {
	k := newTestSimpleK8s()
	ca, cert, key := genSecretData()
	ctx := contextWithDeadline(t)
	k.SaveCertsToSecret(ctx, testSecretName, testNamespace, "cert", "key", ca, cert, key)
	retrievedCert := k.GetCaFromSecret(ctx, testSecretName, testNamespace)
	if !bytes.Equal(retrievedCert, ca) {
		t.Error("Was not able to retrieve CA information that was saved")
	}
}

func TestPatchWebhookConfigurations(t *testing.T) {
	ca, _, _ := genSecretData()

	k := newTestSimpleK8s(
		&admissionv1.MutatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: testWebhookName,
			},
			Webhooks: []admissionv1.MutatingWebhook{{Name: "m1"}, {Name: "m2"}},
		},
		&admissionv1.ValidatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: testWebhookName,
			},
			Webhooks: []admissionv1.ValidatingWebhook{{Name: "v1"}, {Name: "v2"}},
		},
	)

	ctx := contextWithDeadline(t)

	if err := k.patchWebhookConfigurations(ctx, testWebhookName, ca, fail, true, true); err != nil {
		t.Fatalf("Unexpected error patching webhooks: %s: %v", err.Error(), errors.Unwrap(err))
	}

	whmut, err := k.clientset.
		AdmissionregistrationV1().
		MutatingWebhookConfigurations().
		Get(ctx, testWebhookName, metav1.GetOptions{})
	if err != nil {
		t.Error(err)
	}

	whval, err := k.clientset.
		AdmissionregistrationV1().
		MutatingWebhookConfigurations().
		Get(ctx, testWebhookName, metav1.GetOptions{})
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(whmut.Webhooks[0].ClientConfig.CABundle, ca) {
		t.Error("Ca retrieved from first mutating webhook configuration does not match")
	}
	if !bytes.Equal(whmut.Webhooks[1].ClientConfig.CABundle, ca) {
		t.Error("Ca retrieved from second mutating webhook configuration does not match")
	}
	if !bytes.Equal(whval.Webhooks[0].ClientConfig.CABundle, ca) {
		t.Error("Ca retrieved from first validating webhook configuration does not match")
	}
	if !bytes.Equal(whval.Webhooks[1].ClientConfig.CABundle, ca) {
		t.Error("Ca retrieved from second validating webhook configuration does not match")
	}
	if whmut.Webhooks[0].FailurePolicy == nil {
		t.Errorf("Expected first mutating webhook failure policy to be set to %s", fail)
	}
	if whmut.Webhooks[1].FailurePolicy == nil {
		t.Errorf("Expected second mutating webhook failure policy to be set to %s", fail)
	}
	if whval.Webhooks[0].FailurePolicy == nil {
		t.Errorf("Expected first validating webhook failure policy to be set to %s", fail)
	}
	if whval.Webhooks[1].FailurePolicy == nil {
		t.Errorf("Expected second validating webhook failure policy to be set to %s", fail)
	}
}

func Test_Patching_objects(t *testing.T) {
	t.Parallel()

	ctx := contextWithDeadline(t)

	t.Run("returns_error_when", func(t *testing.T) {
		t.Parallel()

		t.Run("failure_policy_is_defined_but_no_webhooks_will_be_patched", func(t *testing.T) {
			t.Parallel()

			k := testK8sWithUnpatchedObjects()

			o := PatchOptions{
				FailurePolicyType: admissionv1.Fail,
			}

			if err := k.PatchObjects(ctx, o); err == nil {
				t.Fatalf("Expected error while patching")
			}
		})

		// This is to preserve old behavior and log format, it could be improved.
		t.Run("diffent_non_empty_names_are_specified_for_validating_and_mutating_webhook", func(t *testing.T) {
			t.Parallel()

			k := testK8sWithUnpatchedObjects()

			o := PatchOptions{
				ValidatingWebhookConfigurationName: "foo",
				MutatingWebhookConfigurationName:   "bar",
			}

			if err := k.PatchObjects(ctx, o); err == nil {
				t.Fatalf("Expected error while patching")
			}
		})

		t.Run("patching_webhook_is_requested_and_it_does_not_exist", func(t *testing.T) {
			t.Parallel()

			k := newTestSimpleK8s()

			o := PatchOptions{
				ValidatingWebhookConfigurationName: "foo",
			}

			if err := k.PatchObjects(ctx, o); err == nil {
				t.Fatalf("Expected error while patching")
			}
		})

		t.Run("patching_APIService_is_requested_and_it_does_not_exist", func(t *testing.T) {
			t.Parallel()

			k := newTestSimpleK8s()

			o := PatchOptions{
				APIServiceName: "foo",
			}

			if err := k.PatchObjects(ctx, o); err == nil {
				t.Fatalf("Expected error while patching")
			}
		})
	})

	t.Run("when_patching_APIService_object", func(t *testing.T) {
		t.Parallel()

		k := testK8sWithUnpatchedObjects()

		o := PatchOptions{
			APIServiceName: testAPIServiceName,
			CABundle:       []byte("foo"),
		}

		if err := k.PatchObjects(ctx, o); err != nil {
			t.Fatalf("Unexpected error while patching objects: %v", err)
		}

		c := k.aggregatorClientset.ApiregistrationV1().APIServices()
		apiService, err := c.Get(ctx, testAPIServiceName, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Unexpected error while getting APIService object %q: %v", testAPIServiceName, err)
		}

		// This is required when CABundle field is populated.
		t.Run("sets_insecure_skip_tls_verity_to_false", func(t *testing.T) {
			t.Parallel()

			if apiService.Spec.InsecureSkipTLSVerify {
				t.Fatalf("Expected insecureSkipTLSVerify of APIService to be false")
			}
		})

		t.Run("sets_ca_bundle_with_ca_certificate_from_created_secret", func(t *testing.T) {
			t.Parallel()

			if len(apiService.Spec.CABundle) == 0 {
				t.Fatalf("Expected CABundle of APIService to be not empty")
			}

			if !bytes.Equal(o.CABundle, apiService.Spec.CABundle) {
				t.Fatalf("CABundle content of APIService does not match requested bundle")
			}
		})
	})

	t.Run("allows_patching_only_validating_webhook", func(t *testing.T) {
		t.Parallel()

		k := testK8sWithUnpatchedObjects()

		o := PatchOptions{
			ValidatingWebhookConfigurationName: testWebhookName,
		}

		if err := k.PatchObjects(ctx, o); err != nil {
			t.Fatalf("Unexpected error patching objects: %v", err)
		}
	})

	t.Run("allows_patching_only_mutating_webhook", func(t *testing.T) {
		t.Parallel()

		k := testK8sWithUnpatchedObjects()

		o := PatchOptions{
			MutatingWebhookConfigurationName: testWebhookName,
		}

		if err := k.PatchObjects(ctx, o); err != nil {
			t.Fatalf("Unexpected error patching objects: %v", err)
		}
	})
}

const (
	// Arbitrary amount of time to let tests exit cleanly before main process terminates.
	timeoutGracePeriod = 10 * time.Second
)

// contextWithDeadline returns context with will timeout before t.Deadline().
func contextWithDeadline(t *testing.T) context.Context {
	t.Helper()

	deadline, ok := t.Deadline()
	if !ok {
		return context.Background()
	}

	ctx, cancel := context.WithDeadline(context.Background(), deadline.Truncate(timeoutGracePeriod))

	t.Cleanup(cancel)

	return ctx
}

func testK8sWithUnpatchedObjects() *k8s {
	ca, cert, key := genSecretData()

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testSecretName,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{"ca": ca, "cert": cert, "key": key},
	}

	validatingWebhook := &admissionv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: testWebhookName,
		},
		Webhooks: []admissionv1.ValidatingWebhook{{Name: "v1"}, {Name: "v2"}},
	}
	mutatingWebhook := &admissionv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: testWebhookName,
		},
		Webhooks: []admissionv1.MutatingWebhook{{Name: "m1"}, {Name: "m2"}},
	}

	apiService := &apiregistrationv1.APIService{
		ObjectMeta: metav1.ObjectMeta{
			Name: testAPIServiceName,
		},
	}

	return &k8s{
		clientset:           fake.NewSimpleClientset(secret, validatingWebhook, mutatingWebhook),
		aggregatorClientset: aggregatorfake.NewSimpleClientset(apiService),
	}
}
