# Custom NGINX template

The NGINX template is located in the file `/etc/nginx/template/nginx.tmpl`.

Using a [Volume](https://kubernetes.io/docs/concepts/storage/volumes/) it is possible to use a custom template.
This includes using a [Configmap](https://kubernetes.io/docs/concepts/storage/volumes/#example-pod-with-a-secret-a-downward-api-and-a-configmap) as source of the template

```yaml
        volumeMounts:
          - mountPath: /etc/nginx/template
            name: nginx-template-volume
            readOnly: true
      volumes:
        - name: nginx-template-volume
          configMap:
            name: nginx-template
            items:
            - key: nginx.tmpl
              path: nginx.tmpl
```

**Please note the template is tied to the Go code. Do not change names in the variable `$cfg`.**

For more information about the template syntax please check the [Go template package](https://golang.org/pkg/text/template/).
In addition to the built-in functions provided by the Go package the following functions are also available:

- empty: returns true if the specified parameter (string) is empty
- contains: [strings.Contains](https://golang.org/pkg/strings/#Contains)
- hasPrefix: [strings.HasPrefix](https://golang.org/pkg/strings/#HasPrefix)
- hasSuffix: [strings.HasSuffix](https://golang.org/pkg/strings/#HasSuffix)
- toUpper: [strings.ToUpper](https://golang.org/pkg/strings/#ToUpper)
- toLower: [strings.ToLower](https://golang.org/pkg/strings/#ToLower)
- quote: wraps a string in double quotes
- buildLocation: helps to build the NGINX Location section in each server
- buildProxyPass: builds the reverse proxy configuration
- buildRateLimit: helps to build a limit zone inside a location if contains a rate limit annotation

TODO:

- buildAuthLocation:
- buildAuthResponseHeaders:
- buildResolvers:
- buildDenyVariable:
- buildUpstreamName:
- buildForwardedFor:
- buildAuthSignURL:
- buildNextUpstream:
- filterRateLimits:
- formatIP:
- getenv:
- getIngressInformation:
- serverConfig:
- isLocationAllowed:
- isValidClientBodyBufferSize:
