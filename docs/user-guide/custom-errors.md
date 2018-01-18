# Custom errors

In case of an error in a request the body of the response is obtained from the `default backend`.
Each request to the default backend includes two headers:

- `X-Code` indicates the HTTP code to be returned to the client.
- `X-Format` the value of the `Accept` header.

**Important:** the custom backend must return the correct HTTP status code to be returned. NGINX do not changes the response from the custom default backend.

Using this two headers is possible to use a custom backend service like [this one](https://github.com/kubernetes/ingress-nginx/tree/master/images/custom-error-pages) that inspect each request and returns a custom error page with the format expected by the client. Please check the example [custom-errors](https://github.com/kubernetes/ingress-nginx/tree/master/docs/examples/customization/custom-errors)

NGINX sends additional headers that can be used to build custom response:

- X-Original-URI
- X-Namespace
- X-Ingress-Name
- X-Service-Name
