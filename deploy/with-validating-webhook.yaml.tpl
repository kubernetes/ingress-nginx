---
apiVersion: v1
kind: Service
metadata:
  name: ingress-validation-webhook
  namespace: ingress-nginx
spec:
  ports:
  - name: admission
    port: 443
    protocol: TCP
    targetPort: 8080
  selector:
    app.kubernetes.io/name: ingress-nginx
---
apiVersion: v1
data:
  key.pem: <key.pem | base64>
  certificate.pem: <certificate.pem | base64>
kind: Secret
metadata:
  name: nginx-ingress-webhook-certificate
  namespace: ingress-nginx
type: Opaque
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-ingress-controller
  namespace: ingress-nginx
  labels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: ingress-nginx
      app.kubernetes.io/part-of: ingress-nginx
  template:
    metadata:
      labels:
        app.kubernetes.io/name: ingress-nginx
        app.kubernetes.io/part-of: ingress-nginx
      annotations:
        prometheus.io/port: "10254"
        prometheus.io/scrape: "true"
    spec:
      serviceAccountName: nginx-ingress-serviceaccount
      containers:
        - name: nginx-ingress-controller
          image: containers.schibsted.io/thibault-jamet/ingress-nginx:0.23.0-schibsted
          args:
            - /nginx-ingress-controller
            - --configmap=$(POD_NAMESPACE)/nginx-configuration
            - --tcp-services-configmap=$(POD_NAMESPACE)/tcp-services
            - --udp-services-configmap=$(POD_NAMESPACE)/udp-services
            - --publish-service=$(POD_NAMESPACE)/ingress-nginx
            - --annotations-prefix=nginx.ingress.kubernetes.io
            - --validating-webhook=:8080
            - --validating-webhook-certificate=/usr/local/certificates/certificate.pem
            - --validating-webhook-key=/usr/local/certificates/key.pem
          volumeMounts:
          - name: webhook-cert
            mountPath: "/usr/local/certificates/"
            readOnly: true
          securityContext:
            allowPrivilegeEscalation: true
            capabilities:
              drop:
                - ALL
              add:
                - NET_BIND_SERVICE
            # www-data -> 33
            runAsUser: 33
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
          ports:
            - name: http
              containerPort: 80
            - name: https
              containerPort: 443
            - name: webhook
              containerPort: 8080
          livenessProbe:
            failureThreshold: 3
            httpGet:
              path: /healthz
              port: 10254
              scheme: HTTP
            initialDelaySeconds: 10
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 10
          readinessProbe:
            failureThreshold: 3
            httpGet:
              path: /healthz
              port: 10254
              scheme: HTTP
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 10
      volumes:
      - name: webhook-cert
        secret:
          secretName: nginx-ingress-webhook-certificate
---
