package k8s

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
)

type k8s struct {
	clientset           kubernetes.Interface
	aggregatorClientset clientset.Interface
}

func New(clientset kubernetes.Interface, aggregatorClientset clientset.Interface) *k8s {
	if clientset == nil {
		log.Fatal("no kubernetes client given")
	}

	if aggregatorClientset == nil {
		log.Fatal("no kubernetes aggregator client given")
	}

	return &k8s{
		clientset:           clientset,
		aggregatorClientset: aggregatorClientset,
	}
}

type PatchOptions struct {
	ValidatingWebhookConfigurationName string
	MutatingWebhookConfigurationName   string
	APIServiceName                     string
	CABundle                           []byte
	FailurePolicyType                  admissionv1.FailurePolicyType
}

func (k8s *k8s) PatchObjects(ctx context.Context, options PatchOptions) error {
	patchMutating := options.MutatingWebhookConfigurationName != ""
	patchValidating := options.ValidatingWebhookConfigurationName != ""
	patchAPIService := options.APIServiceName != ""

	if !patchMutating && !patchValidating && options.FailurePolicyType != "" {
		return fmt.Errorf("failurePolicy specified, but no webhook will be patched")
	}

	if patchMutating && patchValidating &&
		options.MutatingWebhookConfigurationName != options.ValidatingWebhookConfigurationName {
		return fmt.Errorf("webhook names must be the same")
	}

	if patchAPIService {
		log.Infof("patching APIService %q", options.APIServiceName)

		if err := k8s.patchAPIService(ctx, options.APIServiceName, options.CABundle); err != nil {
			// Intentionally don't wrap error here to preserve old behavior and be able to log both
			// original error and a message.
			return err
		}
	}

	webhookName := options.ValidatingWebhookConfigurationName
	if webhookName == "" {
		webhookName = options.MutatingWebhookConfigurationName
	}

	if patchMutating || patchValidating {
		return k8s.patchWebhookConfigurations(ctx, webhookName, options.CABundle, options.FailurePolicyType, patchMutating, patchValidating)
	}

	return nil
}

func (k8s *k8s) patchAPIService(ctx context.Context, objectName string, ca []byte) error {
	log.Infof("patching APIService %q", objectName)

	c := k8s.aggregatorClientset.ApiregistrationV1().APIServices()

	apiService, err := c.Get(ctx, objectName, metav1.GetOptions{})
	if err != nil {
		return &wrappedError{
			err:     err,
			message: fmt.Sprintf("failed getting APIService %q", objectName),
		}
	}

	apiService.Spec.CABundle = ca
	apiService.Spec.InsecureSkipTLSVerify = false

	if _, err := c.Update(ctx, apiService, metav1.UpdateOptions{}); err != nil {
		return &wrappedError{
			err:     err,
			message: fmt.Sprintf("failed patching APIService %q", objectName),
		}
	}

	log.Debug("patched APIService")

	return nil
}

// patchWebhookConfigurations will patch validatingWebhook and mutatingWebhook clientConfig configurations with
// the provided ca data. If failurePolicy is provided, patch all webhooks with this value
func (k8s *k8s) patchWebhookConfigurations(
	ctx context.Context,
	configurationName string,
	ca []byte,
	failurePolicy admissionv1.FailurePolicyType,
	patchMutating bool,
	patchValidating bool,
) error {
	log.Infof("patching webhook configurations '%s' mutating=%t, validating=%t, failurePolicy=%s", configurationName, patchMutating, patchValidating, failurePolicy)

	if patchValidating {
		if err := k8s.patchValidating(ctx, configurationName, ca, failurePolicy); err != nil {
			// Intentionally don't wrap error here to preserve old behavior and be able to log both original error and a message.
			return err
		}
	} else {
		log.Debug("validating hook patching not required")
	}

	if patchMutating {
		if err := k8s.patchMutating(ctx, configurationName, ca, failurePolicy); err != nil {
			// Intentionally don't wrap error here to preserve old behavior and be able to log both original error and a message.
			return err
		}
	} else {
		log.Debug("mutating hook patching not required")
	}

	log.Info("Patched hook(s)")

	return nil
}

type wrappedError struct {
	err     error
	message string
}

func (err wrappedError) Error() string {
	return err.message
}

func (err wrappedError) Unwrap() error {
	return err.err
}

func (k8s *k8s) patchValidating(ctx context.Context, configurationName string, ca []byte, failurePolicy admissionv1.FailurePolicyType) error {
	valHook, err := k8s.clientset.
		AdmissionregistrationV1().
		ValidatingWebhookConfigurations().
		Get(ctx, configurationName, metav1.GetOptions{})
	if err != nil {
		return &wrappedError{
			err:     err,
			message: "failed getting validating webhook",
		}
	}

	for i := range valHook.Webhooks {
		h := &valHook.Webhooks[i]
		h.ClientConfig.CABundle = ca
		if failurePolicy != "" {
			h.FailurePolicy = &failurePolicy
		}
	}

	if _, err = k8s.clientset.AdmissionregistrationV1().
		ValidatingWebhookConfigurations().
		Update(ctx, valHook, metav1.UpdateOptions{}); err != nil {
		return &wrappedError{
			err:     err,
			message: "failed patching validating webhook",
		}
	}
	log.Debug("patched validating hook")

	return nil
}

func (k8s *k8s) patchMutating(ctx context.Context, configurationName string, ca []byte, failurePolicy admissionv1.FailurePolicyType) error {
	mutHook, err := k8s.clientset.
		AdmissionregistrationV1().
		MutatingWebhookConfigurations().
		Get(ctx, configurationName, metav1.GetOptions{})
	if err != nil {
		return &wrappedError{
			err:     err,
			message: "failed getting mutating webhook",
		}
	}

	for i := range mutHook.Webhooks {
		h := &mutHook.Webhooks[i]
		h.ClientConfig.CABundle = ca
		if failurePolicy != "" {
			h.FailurePolicy = &failurePolicy
		}
	}

	if _, err = k8s.clientset.AdmissionregistrationV1().
		MutatingWebhookConfigurations().
		Update(ctx, mutHook, metav1.UpdateOptions{}); err != nil {
		return &wrappedError{
			err:     err,
			message: "failed patching mutating webhook",
		}
	}
	log.Debug("patched mutating hook")

	return nil
}

// GetCaFromSecret will check for the presence of a secret. If it exists, will return the content of the
// "ca" from the secret, otherwise will return nil
func (k8s *k8s) GetCaFromSecret(ctx context.Context, secretName string, namespace string) []byte {
	log.Debugf("getting secret '%s' in namespace '%s'", secretName, namespace)
	secret, err := k8s.clientset.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
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
func (k8s *k8s) SaveCertsToSecret(ctx context.Context, secretName, namespace, certName, keyName string, ca, cert, key []byte) {
	log.Debugf("saving to secret '%s' in namespace '%s'", secretName, namespace)
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{"ca": ca, certName: cert, keyName: key},
	}

	log.Debug("saving secret")
	_, err := k8s.clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		log.WithField("err", err).Fatal("failed creating secret")
	}
	log.Debug("saved secret")
}
