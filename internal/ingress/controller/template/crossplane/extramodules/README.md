# Extra modules
This folder contains the extra modules used by ingress-nginx and not yet 
supported by nginx-go-crossplane

The generation of the files is done using go-crossplane generator

## Brotli
```
go run ./cmd/generate/ -src-path=ngx_brotli/ -directive-map-name=brotliDirectives -match-func-name=BrotliMatchFn > ../ingress-nginx/internal/ingress/controller/template/crossplane/extramodules/brotli.go
```