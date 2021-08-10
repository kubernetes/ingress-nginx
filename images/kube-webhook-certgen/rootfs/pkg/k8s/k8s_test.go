package k8s

import (
	"bytes"
	"context"
	"math/rand"
	"testing"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	testWebhookName = "c7c95710-d8c3-4cc3-a2a8-8d2b46909c76"
	testSecretName  = "15906410-af2a-4f9b-8a2d-c08ffdd5e129"
	testNamespace   = "7cad5f92-c0d5-4bc9-87a3-6f44d5a5619d"
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

func newTestSimpleK8s() *k8s {
	return &k8s{
		clientset: fake.NewSimpleClientset(),
	}
}

func TestGetCaFromCertificate(t *testing.T) {
	k := newTestSimpleK8s()

	ca, cert, key := genSecretData()

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: testSecretName,
		},
		Data: map[string][]byte{"ca": ca, "cert": cert, "key": key},
	}

	k.clientset.CoreV1().Secrets(testNamespace).Create(context.Background(), secret, metav1.CreateOptions{})

	retrievedCa := k.GetCaFromSecret(testSecretName, testNamespace)
	if !bytes.Equal(retrievedCa, ca) {
		t.Error("Was not able to retrieve CA information that was saved")
	}
}

func TestSaveCertsToSecret(t *testing.T) {
	k := newTestSimpleK8s()

	ca, cert, key := genSecretData()

	k.SaveCertsToSecret(testSecretName, testNamespace, "cert", "key", ca, cert, key)

	secret, _ := k.clientset.CoreV1().Secrets(testNamespace).Get(context.Background(), testSecretName, metav1.GetOptions{})

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
	k.SaveCertsToSecret(testSecretName, testNamespace, "cert", "key", ca, cert, key)
	retrievedCert := k.GetCaFromSecret(testSecretName, testNamespace)
	if !bytes.Equal(retrievedCert, ca) {
		t.Error("Was not able to retrieve CA information that was saved")
	}
}

func TestPatchWebhookConfigurations(t *testing.T) {
	k := newTestSimpleK8s()

	ca, _, _ := genSecretData()

	k.clientset.
		AdmissionregistrationV1().
		MutatingWebhookConfigurations().
		Create(context.Background(), &admissionv1.MutatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: testWebhookName,
			},
			Webhooks: []admissionv1.MutatingWebhook{{Name: "m1"}, {Name: "m2"}}}, metav1.CreateOptions{})

	k.clientset.
		AdmissionregistrationV1().
		ValidatingWebhookConfigurations().
		Create(context.Background(), &admissionv1.ValidatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: testWebhookName,
			},
			Webhooks: []admissionv1.ValidatingWebhook{{Name: "v1"}, {Name: "v2"}}}, metav1.CreateOptions{})

	k.PatchWebhookConfigurations(testWebhookName, ca, &fail, true, true)

	whmut, err := k.clientset.
		AdmissionregistrationV1().
		MutatingWebhookConfigurations().
		Get(context.Background(), testWebhookName, metav1.GetOptions{})

	if err != nil {
		t.Error(err)
	}

	whval, err := k.clientset.
		AdmissionregistrationV1beta1().
		MutatingWebhookConfigurations().
		Get(context.Background(), testWebhookName, metav1.GetOptions{})

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
