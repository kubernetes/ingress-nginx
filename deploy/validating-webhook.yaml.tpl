---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: check-ingress
webhooks:
- name: validate.nginx.ingress.kubernetes.io
  rules:
  - apiGroups:
    - extensions
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - ingresses
  failurePolicy: Fail
  clientConfig:
    service:
      namespace: ingress-nginx
      name: nginx-ingress-webhook
      path: /extensions/v1beta1/ingresses
    caBundle: <certificate.pem | base64>
---