package cmd_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/jet/kube-webhook-certgen/cmd"
	"github.com/jet/kube-webhook-certgen/pkg/k8s"
)

func Test_Patch(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	t.Run("patches_APIService_object_when_requested", func(t *testing.T) {
		t.Parallel()

		config := testPatchConfig()
		config.APIServiceName = "bar"

		patcher := testPatcher()
		patcher.patchObjects = func(_ context.Context, options k8s.PatchOptions) error {
			if options.APIServiceName != config.APIServiceName {
				return fmt.Errorf("unexpected APIService name %q, expected %q", options.APIServiceName, config.APIServiceName)
			}

			return nil
		}
		config.Patcher = patcher

		if err := cmd.Patch(ctx, config); err != nil {
			t.Fatalf("Unexpected patching error: %v", err)
		}
	})

	t.Run("use_configured_webhook_name_for_patching", func(t *testing.T) {
		t.Parallel()

		config := testPatchConfig()
		config.WebhookName = "foo"

		patcher := testPatcher()
		patcher.patchObjects = func(_ context.Context, options k8s.PatchOptions) error {
			if options.ValidatingWebhookConfigurationName != config.WebhookName {
				return fmt.Errorf("unexpected object name %q, expected %q", options.ValidatingWebhookConfigurationName, config.WebhookName)
			}

			return nil
		}
		config.Patcher = patcher

		if err := cmd.Patch(ctx, config); err != nil {
			t.Fatalf("Unexpected patching error: %v", err)
		}
	})

	t.Run("patches_only_validating_webhook_when_requested", func(t *testing.T) {
		t.Parallel()

		config := testPatchConfig()
		config.PatchValidating = true
		config.PatchMutating = false

		patcher := testPatcher()
		patcher.patchObjects = func(_ context.Context, options k8s.PatchOptions) error {
			if options.ValidatingWebhookConfigurationName == "" {
				t.Error("expected validating webhook to be patched")
			}

			if options.MutatingWebhookConfigurationName != "" {
				t.Error("expected mutating webhook to not be patched")
			}

			if options.APIServiceName != "" {
				t.Error("expected APIService to not be patched")
			}

			return nil
		}
		config.Patcher = patcher

		if err := cmd.Patch(ctx, config); err != nil {
			t.Fatalf("Unexpected patching error: %v", err)
		}
	})

	t.Run("patches_both_webhooks_when_requested", func(t *testing.T) {
		t.Parallel()

		config := testPatchConfig()
		config.PatchValidating = true
		config.PatchMutating = true

		patcher := testPatcher()
		patcher.patchObjects = func(_ context.Context, options k8s.PatchOptions) error {
			if options.ValidatingWebhookConfigurationName == "" {
				t.Error("expected validating webhook to be patched")
			}

			if options.MutatingWebhookConfigurationName == "" {
				t.Error("expected mutating webhook to be patched")
			}

			if options.APIServiceName != "" {
				t.Error("expected APIService to not be patched")
			}

			return nil
		}
		config.Patcher = patcher

		if err := cmd.Patch(ctx, config); err != nil {
			t.Fatalf("Unexpected patching error: %v", err)
		}
	})

	t.Run("use_empty_policy_when_ignore_is_requested", func(t *testing.T) {
		t.Parallel()

		config := testPatchConfig()
		config.PatchFailurePolicy = "Ignore"

		patcher := testPatcher()
		patcher.patchObjects = func(_ context.Context, options k8s.PatchOptions) error {
			if options.FailurePolicyType != "" {
				return fmt.Errorf("expected policy to be nil. got: %q", options.FailurePolicyType)
			}

			return nil
		}
		config.Patcher = patcher

		if err := cmd.Patch(ctx, config); err != nil {
			t.Fatalf("Unexpected patching error: %v", err)
		}
	})

	t.Run("use_fail_policy_when_fail_is_requested", func(t *testing.T) {
		t.Parallel()

		config := testPatchConfig()
		config.PatchFailurePolicy = "Fail"

		patcher := testPatcher()
		patcher.patchObjects = func(_ context.Context, options k8s.PatchOptions) error {
			if options.FailurePolicyType == "" || options.FailurePolicyType != "Fail" {
				return fmt.Errorf("unexpected policy: %q", options.FailurePolicyType)
			}

			return nil
		}
		config.Patcher = patcher

		if err := cmd.Patch(ctx, config); err != nil {
			t.Fatalf("Unexpected patching error: %v", err)
		}
	})

	t.Run("use_obtained_ca_certificate_for_patching", func(t *testing.T) {
		t.Parallel()

		expectedCA := []byte("foo")

		config := testPatchConfig()

		patcher := testPatcher()
		patcher.patchObjects = func(_ context.Context, options k8s.PatchOptions) error {
			if !reflect.DeepEqual(options.CABundle, expectedCA) {
				return fmt.Errorf("unexpected CA, expected %q, got %q", string(expectedCA), string(options.CABundle))
			}

			return nil
		}
		patcher.getCaFromSecret = func(context.Context, string, string) []byte {
			return expectedCA
		}
		config.Patcher = patcher

		if err := cmd.Patch(ctx, config); err != nil {
			t.Fatalf("Unexpected patching error: %v", err)
		}
	})

	t.Run("returns_error_when", func(t *testing.T) {
		t.Parallel()

		for name, mutateF := range map[string]func(*cmd.PatchConfig){
			"no_patcher_is_defined": func(c *cmd.PatchConfig) {
				c.Patcher = nil
			},
			"no_webhooks_are_requested_for_patching": func(c *cmd.PatchConfig) {
				c.PatchValidating = false
				c.PatchMutating = false
				c.APIServiceName = ""
			},
			"unsupported_patch_failure_policy_is_defined": func(c *cmd.PatchConfig) {
				c.PatchFailurePolicy = "foo"
			},
			"ca_certificate_from_secret_is_empty": func(c *cmd.PatchConfig) {
				patcher := testPatcher()
				patcher.getCaFromSecret = func(_ context.Context, _, _ string) []byte {
					return nil
				}
				c.Patcher = patcher
			},
		} {
			mutateF := mutateF

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				config := testPatchConfig()
				mutateF(config)

				if err := cmd.Patch(ctx, config); err == nil {
					t.Fatalf("Expected error while patching")
				}
			})
		}
	})
}

type patcher struct {
	patchObjects    func(context.Context, k8s.PatchOptions) error
	getCaFromSecret func(context.Context, string, string) []byte
}

func (p *patcher) PatchObjects(ctx context.Context, options k8s.PatchOptions) error {
	return p.patchObjects(ctx, options)
}

func (p *patcher) GetCaFromSecret(ctx context.Context, secretName, namespace string) []byte {
	return p.getCaFromSecret(ctx, secretName, namespace)
}

func testPatcher() *patcher {
	return &patcher{
		patchObjects: func(context.Context, k8s.PatchOptions) error {
			return nil
		},
		getCaFromSecret: func(context.Context, string, string) []byte { return []byte{} },
	}
}

func testPatchConfig() *cmd.PatchConfig {
	return &cmd.PatchConfig{
		PatchValidating: true,
		WebhookName:     "foo",
		Patcher:         testPatcher(),
	}
}
