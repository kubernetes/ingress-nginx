package cmd

import (
	"os"

	"github.com/jet/kube-webhook-certgen/pkg/k8s"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	admissionv1 "k8s.io/api/admissionregistration/v1"
)

var (
	patch = &cobra.Command{
		Use:    "patch",
		Short:  "Patch a validatingwebhookconfiguration and mutatingwebhookconfiguration 'webhook-name' by using the ca from 'secret-name' in 'namespace'",
		Long:   "Patch a validatingwebhookconfiguration and mutatingwebhookconfiguration 'webhook-name' by using the ca from 'secret-name' in 'namespace'",
		PreRun: prePatchCommand,
		Run:    patchCommand}
)

func prePatchCommand(cmd *cobra.Command, args []string) {
	configureLogging(cmd, args)
	if cfg.patchMutating == false && cfg.patchValidating == false {
		log.Fatal("patch-validating=false, patch-mutating=false. You must patch at least one kind of webhook, otherwise this command is a no-op")
		os.Exit(1)
	}
	switch cfg.patchFailurePolicy {
	case "":
		break
	case "Ignore":
	case "Fail":
		failurePolicy = admissionv1.FailurePolicyType(cfg.patchFailurePolicy)
		break
	default:
		log.Fatalf("patch-failure-policy %s is not valid", cfg.patchFailurePolicy)
		os.Exit(1)
	}
}

func patchCommand(_ *cobra.Command, _ []string) {
	k := k8s.New(cfg.kubeconfig)
	ca := k.GetCaFromSecret(cfg.secretName, cfg.namespace)

	if ca == nil {
		log.Fatalf("no secret with '%s' in '%s'", cfg.secretName, cfg.namespace)
	}

	k.PatchWebhookConfigurations(cfg.webhookName, ca, &failurePolicy, cfg.patchMutating, cfg.patchValidating)
}

func init() {
	rootCmd.AddCommand(patch)
	patch.Flags().StringVar(&cfg.secretName, "secret-name", "", "Name of the secret where certificate information will be read from")
	patch.Flags().StringVar(&cfg.namespace, "namespace", "", "Namespace of the secret where certificate information will be read from")
	patch.Flags().StringVar(&cfg.webhookName, "webhook-name", "", "Name of validatingwebhookconfiguration and mutatingwebhookconfiguration that will be updated")
	patch.Flags().BoolVar(&cfg.patchValidating, "patch-validating", true, "If true, patch validatingwebhookconfiguration")
	patch.Flags().BoolVar(&cfg.patchMutating, "patch-mutating", true, "If true, patch mutatingwebhookconfiguration")
	patch.Flags().StringVar(&cfg.patchFailurePolicy, "patch-failure-policy", "", "If set, patch the webhooks with this failure policy. Valid options are Ignore or Fail")
	patch.MarkFlagRequired("secret-name")
	patch.MarkFlagRequired("namespace")
	patch.MarkFlagRequired("webhook-name")
}
