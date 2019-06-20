# NGINX status page

The [ngx_http_stub_status_module](http://nginx.org/en/docs/http/ngx_http_stub_status_module.html) module provides access to basic status information.
This is the default module active in the url `/nginx_status` in the status port (default is 18080).

This controller provides an alternative to this module using the [nginx-module-vts](https://github.com/vozlt/nginx-module-vts) module.
To use this module just set in the configuration configmap `enable-vts-status: "true"`.

![nginx-module-vts screenshot](https://cloud.githubusercontent.com/assets/3648408/10876811/77a67b70-8183-11e5-9924-6a6d0c5dc73a.png "screenshot with filter")

To extract the information in JSON format the module provides a custom URL: `/nginx_status/format/json`
