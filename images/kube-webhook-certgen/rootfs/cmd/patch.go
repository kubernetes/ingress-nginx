package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/jet/kube-webhook-certgen/pkg/k8s"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	admissionv1 "k8s.io/api/admissionregistration/v1"
)

var patch = &cobra.Command{
	Use:    "patch",
	Short:  "Patch a ValidatingWebhookConfiguration, MutatingWebhookConfiguration or APIService 'object-name' by using the ca from 'secret-name' in 'namespace'",
	Long:   "Patch a ValidatingWebhookConfiguration, MutatingWebhookConfiguration or APIService 'object-name' by using the ca from 'secret-name' in 'namespace'",
	PreRun: configureLogging,
	Run:    patchCommand,
}

type PatchConfig struct {
	PatchMutating      bool
	PatchValidating    bool
	PatchFailurePolicy string
	APIServiceName     string
	WebhookName        string

	SecretName string
	Namespace  string

	Patcher Patcher
}

type Patcher interface {
	PatchObjects(ctx context.Context, options k8s.PatchOptions) error
	GetCaFromSecret(ctx context.Context, secretName, namespace string) []byte
}

func Patch(ctx context.Context, cfg *PatchConfig) error {
	if cfg.Patcher == nil {
		return fmt.Errorf("no patcher defined")
	}

	if !cfg.PatchMutating && !cfg.PatchValidating && cfg.APIServiceName == "" {
		return fmt.Errorf("patch-validating=false, patch-mutating=false. You must patch at least one kind of webhook, otherwise this command is a no-op")
	}

	var failurePolicy admissionv1.FailurePolicyType

	switch cfg.PatchFailurePolicy {
	case "":
		break
	case "Ignore":
	case "Fail":
		failurePolicy = admissionv1.FailurePolicyType(cfg.PatchFailurePolicy)
		break
	default:
		return fmt.Errorf("patch-failure-policy %s is not valid", cfg.PatchFailurePolicy)
	}

	ca := cfg.Patcher.GetCaFromSecret(ctx, cfg.SecretName, cfg.Namespace)

	if ca == nil {
		return fmt.Errorf("no secret with '%s' in '%s'", cfg.SecretName, cfg.Namespace)
	}

	options := k8s.PatchOptions{
		CABundle:          ca,
		FailurePolicyType: failurePolicy,
		APIServiceName:    cfg.APIServiceName,
	}

	if cfg.PatchMutating {
		options.MutatingWebhookConfigurationName = cfg.WebhookName
	}

	if cfg.PatchValidating {
		options.ValidatingWebhookConfigurationName = cfg.WebhookName
	}

	return cfg.Patcher.PatchObjects(ctx, options)
}

func patchCommand(_ *cobra.Command, _ []string) {
	client, aggregationClient := newKubernetesClients(cfg.kubeconfig)

	config := &PatchConfig{
		SecretName:         cfg.secretName,
		Namespace:          cfg.namespace,
		PatchMutating:      cfg.patchMutating,
		PatchValidating:    cfg.patchValidating,
		PatchFailurePolicy: cfg.patchFailurePolicy,
		APIServiceName:     cfg.apiServiceName,
		WebhookName:        cfg.webhookName,
		Patcher:            k8s.New(client, aggregationClient),
	}

	ctx := context.TODO()

	if err := Patch(ctx, config); err != nil {
		if wrappedErr := errors.Unwrap(err); wrappedErr != nil {
			log.WithField("err", wrappedErr).Fatal(err.Error())
		}

		log.Fatal(err.Error())
	}
}

func init() {
	rootCmd.AddCommand(patch)
	patch.Flags().StringVar(&cfg.secretName, "secret-name", "", "Name of the secret where certificate information will be read from")
	patch.Flags().StringVar(&cfg.namespace, "namespace", "", "Namespace of the secret where certificate information will be read from")
	patch.Flags().StringVar(&cfg.webhookName, "webhook-name", "", "Name of ValidatingWebhookConfiguration and MutatingWebhookConfiguration that will be updated")
	patch.Flags().StringVar(&cfg.apiServiceName, "apiservice-name", "", "Name of APIService that will be patched")
	patch.Flags().BoolVar(&cfg.patchValidating, "patch-validating", true, "If true, patch ValidatingWebhookConfiguration")
	patch.Flags().BoolVar(&cfg.patchMutating, "patch-mutating", true, "If true, patch MutatingWebhookConfiguration")
	patch.Flags().StringVar(&cfg.patchFailurePolicy, "patch-failure-policy", "", "If set, patch the webhooks with this failure policy. Valid options are Ignore or Fail")
	patch.MarkFlagRequired("secret-name")
	patch.MarkFlagRequired("namespace")
}
