# NGINX Configuration

There are three ways to customize NGINX:

1. [ConfigMap](./configmap.md): using a Configmap to set global configurations in NGINX.
2. [Annotations](./annotations.md): use this if you want a specific configuration for a particular Ingress rule.
3. [Custom template](./custom-template.md): when more specific settings are required, like [open_file_cache](https://nginx.org/en/docs/http/ngx_http_core_module.html#open_file_cache), adjust [listen](https://nginx.org/en/docs/http/ngx_http_core_module.html#listen) options as `rcvbuf` or when is not possible to change the configuration through the ConfigMap.
