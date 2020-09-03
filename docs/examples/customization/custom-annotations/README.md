# Custom annotations

Demonstrates how to use arbitrary ingress annotations, and notably custom ones.

In this example, we wish to conditionally add an ingress-specific request header (only if not already set by client).
This could be done with a Lua block:

```lua
rewrite_by_lua_block {
  headers = ngx.req.get_headers()
  if not headers["foo"] then
    ngx.req.set_header["foo"] = "bar"
  end
}
```

We might try to add it via [configuration snippets](../configuration-snippets/README.md), only to find out that configuration reloads fail due to duplicate `rewrite_by_lua_block` directives, as another one is already present in the default template.

To work around this issue, we must find a way to insert ingress-specific configuration inside the default `rewrite_by_lua_block` directive.

## Plugins: a poor solution

We could write a Lua plugin, using a `ngx.var.host` filter to determine which headers to add in `rewrite()` (akin to the [`hello_world` plugin example](../../../../rootfs/etc/nginx/lua/plugins/hello_world)). However, this separates an ingress definition from its headers, which is inelegant and poorly maintainable.

## Custom annotations: the right solution

We can access raw ingress data in NGINX templates via the already defined `$ing` variable:

```nginx
{{ $ing := (getIngressInformation $location.Ingress $server.Hostname $location.Path) }}
```

This enables us to craft a [custom template](../../../user-guide/nginx-configuration/custom-template.md), where we access custom annotations in the default `rewrite_by_lua_block` directive:

```nginx
rewrite_by_lua_block {
  ... -- default template config

  -- Custom headers
  {{ if index $ing.Annotations "my-annotations/custom-headers" }}
  {{ index $ing.Annotations "my-annotations/custom-headers" }}
  {{ end }}
}
```

It is then possible to insert conditional headers directly in an ingress definition (see [ingress.yaml](ingress.yaml) for full ingress definition):

```yaml
my-annotations/custom-headers: |
  headers = ngx.req.get_headers()
  if not headers["foo"] then
    ngx.req.set_header["foo"] = "bar"
  end
```

Hopefully this gives you some ideas to play with custom annotations :)
