package k8s

import (
	"context"

	log "github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type k8s struct {
	clientset kubernetes.Interface
}

func New(kubeconfig string) *k8s {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.WithError(err).Fatal("error building kubernetes config")
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.WithError(err).Fatal("error creating kubernetes client")
	}

	return &k8s{clientset: c}
}

// PatchWebhookConfigurations will patch validatingWebhook and mutatingWebhook clientConfig configurations with
// the provided ca data. If failurePolicy is provided, patch all webhooks with this value
func (k8s *k8s) PatchWebhookConfigurations(
	configurationNames string, ca []byte,
	failurePolicy *admissionv1.FailurePolicyType,
	patchMutating bool, patchValidating bool) {

	log.Infof("patching webhook configurations '%s' mutating=%t, validating=%t, failurePolicy=%s", configurationNames, patchMutating, patchValidating, *failurePolicy)

	if patchValidating {
		valHook, err := k8s.clientset.
			AdmissionregistrationV1().
			ValidatingWebhookConfigurations().
			Get(context.TODO(), configurationNames, metav1.GetOptions{})

		if err != nil {
			log.WithField("err", err).Fatal("failed getting validating webhook")
		}

		for i := range valHook.Webhooks {
			h := &valHook.Webhooks[i]
			h.ClientConfig.CABundle = ca
			if *failurePolicy != "" {
				h.FailurePolicy = failurePolicy
			}
		}

		if _, err = k8s.clientset.AdmissionregistrationV1().
			ValidatingWebhookConfigurations().
			Update(context.TODO(), valHook, metav1.UpdateOptions{}); err != nil {
			log.WithField("err", err).Fatal("failed patching validating webhook")
		}
		log.Debug("patched validating hook")
	} else {
		log.Debug("validating hook patching not required")
	}

	if patchMutating {
		mutHook, err := k8s.clientset.
			AdmissionregistrationV1().
			MutatingWebhookConfigurations().
			Get(context.TODO(), configurationNames, metav1.GetOptions{})
		if err != nil {
			log.WithField("err", err).Fatal("failed getting validating webhook")
		}

		for i := range mutHook.Webhooks {
			h := &mutHook.Webhooks[i]
			h.ClientConfig.CABundle = ca
			if *failurePolicy != "" {
				h.FailurePolicy = failurePolicy
			}
		}

		if _, err = k8s.clientset.AdmissionregistrationV1().
			MutatingWebhookConfigurations().
			Update(context.TODO(), mutHook, metav1.UpdateOptions{}); err != nil {
			log.WithField("err", err).Fatal("failed patching validating webhook")
		}
		log.Debug("patched mutating hook")
	} else {
		log.Debug("mutating hook patching not required")
	}

	log.Info("Patched hook(s)")
}

// GetCaFromSecret will check for the presence of a secret. If it exists, will return the content of the
// "ca" from the secret, otherwise will return nil
func (k8s *k8s) GetCaFromSecret(secretName string, namespace string) []byte {
	log.Debugf("getting secret '%s' in namespace '%s'", secretName, namespace)
	secret, err := k8s.clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.WithField("err", err).Info("no secret found")
			return nil
		}
		log.WithField("err", err).Fatal("error getting secret")
	}

	data := secret.Data["ca"]
	if data == nil {
		log.Fatal("got secret, but it did not contain a 'ca' key")
	}
	log.Debug("got secret")
	return data
}

// SaveCertsToSecret saves the provided ca, cert and key into a secret in the specified namespace.
func (k8s *k8s) SaveCertsToSecret(secretName, namespace, certName, keyName string, ca, cert, key []byte) {

	log.Debugf("saving to secret '%s' in namespace '%s'", secretName, namespace)
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{"ca": ca, certName: cert, keyName: key},
	}

	log.Debug("saving secret")
	_, err := k8s.clientset.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		log.WithField("err", err).Fatal("failed creating secret")
	}
	log.Debug("saved secret")
}
