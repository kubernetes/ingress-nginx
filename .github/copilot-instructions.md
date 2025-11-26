# AI Assistant Guidelines for `ingress-nginx`

These instructions are for AI coding agents working in this repository.

## Big Picture
- This repo implements the Kubernetes Ingress controller using NGINX as the data plane.
- The Go controller watches Kubernetes objects (Ingress, Service, Endpoints, Secrets, ConfigMaps), builds an internal model, renders `nginx.tmpl` into `nginx.conf`, and coordinates a companion NGINX + Lua container.
- Static config changes typically require an NGINX reload; endpoint and some upstream-only changes are pushed dynamically via Lua without reload.
- Core entrypoints:
  - `cmd/nginx/main.go`: full controller + API client + metrics HTTP server.
  - `cmd/dataplane/main.go`: dataplane-only variant sharing the same controller package.

## Code Structure & Key Packages
- `internal/ingress/controller`: reconciliation loop, sync queue, and NGINX controller orchestration. Prefer extending behavior here rather than duplicating logic elsewhere.
- `internal/ingress/annotations`: defines supported annotations and parsing. New annotation behaviors should be added here and wired into the model.
- `internal/ingress/types.go`: core model types used to render the NGINX template.
- `internal/nginx`: helpers for checking NGINX state, profiler/health endpoints.
- `internal/net`, `internal/k8s`: networking and Kubernetes helpers; reuse instead of inlining parsing/lookup logic.
- `rootfs/etc/nginx/template/nginx.tmpl`: Go template used to generate `nginx.conf` from the model.
- `rootfs/etc/nginx/lua/**`: Lua code for dynamic endpoints, canary, auth, etc.; changes here often pair with Go changes in `internal/ingress`.
- `test/e2e/**`: Ginkgo-based end-to-end tests organized by feature (e.g., default backend, annotations).
- `charts/ingress-nginx/**`: Helm chart and its own `README.md` & changelog; keep chart values, docs, and controller flags in sync.

## Development & Build Workflow
- Most developer workflows are driven via `make` and Dockerized tooling (see `Makefile`).
- Local dev cluster and controller deployment:
  - `make dev-env` — creates a kind cluster, builds the image, and deploys ingress-nginx (primary way to iterate on features).
  - `make dev-env-stop` — tears down the dev cluster.
- Building and images:
  - `make build` — builds controller binaries inside Docker using `build/run-in-docker.sh`.
  - `make image` or `make image-chroot` — builds controller images for the configured `ARCH` and `REGISTRY`.
  - For custom images, follow the pattern in `docs/developer-guide/getting-started.md` (e.g., set `TAG` and `REGISTRY`, then `make build image`).

## Testing Conventions
- Go unit tests:
  - `make test` — runs Go unit tests inside the standard build container.
  - Use existing patterns under `internal/**` and `pkg/**`; prefer table-driven tests.
- Lua unit tests:
  - `make lua-test` — runs tests under `rootfs/etc/nginx/lua/test`.
  - Test files must end with `_test.lua` or they are ignored.
- E2E tests:
  - `make kind-e2e-test` — runs the main suite against a kind cluster.
  - Restrict scope with `FOCUS="<ginkgo-focus>" make kind-e2e-test` (see `docs/e2e-tests.md` for available focus labels).
- Helm tests:
  - `make helm-test` — runs helm-unittest tests in `charts/ingress-nginx/tests`.

## Project-Specific Patterns & Rules
- **Ingress model diffing**: controller maintains current and desired models and diffs them; honor this pattern by updating model-building code instead of triggering reloads directly.
- **Reload minimization**: endpoint and upstream-only changes should flow through the Lua-based dynamic update path; avoid introducing new changes that require full NGINX reload when not strictly necessary.
- **Ordering & conflict resolution**:
  - Ingress rules are ordered by `CreationTimestamp`; the oldest rule "wins" for duplicate host/path or TLS sections.
  - Annotations apply per-Ingress; do not assume cross-Ingress sharing of annotation state.
- **Annotations**: when adding or changing annotation behavior, update:
  - Parsing in `internal/ingress/annotations/**`.
  - Model wiring in `internal/ingress/controller` or related helpers.
  - Relevant docs in `docs/user-guide/nginx-configuration/annotations.md` and examples under `docs/examples/**`.
- **Multi-component alignment**: many features span Go controller, NGINX template, Lua, and docs; search across `internal/ingress`, `rootfs/etc/nginx`, and `docs/**` before altering behavior.

## Documentation & Website
- Docs are in `docs/**` and rendered via MkDocs (`Makefile` target `live-docs` and `build-docs`).
- When changing flags, annotations, or public behavior, update:
  - Top-level `README.md` (support matrix or high-level messaging, if impacted).
  - Relevant pages under `docs/user-guide/**` and `docs/examples/**`.
  - Helm chart docs in `charts/ingress-nginx/README.md` when chart values or defaults change.

## How to Ask the AI for Changes
- When requesting changes, reference concrete paths and packages (e.g., `internal/ingress/annotations/rewrite/main.go`, `rootfs/etc/nginx/template/nginx.tmpl`).
- For behavior changes that affect traffic routing, TLS, or auth, also specify how they should be covered by tests (`test/e2e/**` and/or Lua tests).
- Avoid introducing new CLIs, build systems, or frameworks; extend existing make targets, controller flags, and Helm values instead.
