# Ingress NGINX Controller

[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes/ingress-nginx)](https://goreportcard.com/report/github.com/kubernetes/ingress-nginx)
[![GitHub license](https://img.shields.io/github/license/kubernetes/ingress-nginx.svg)](https://github.com/kubernetes/ingress-nginx/blob/main/LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/kubernetes/ingress-nginx.svg)](https://github.com/kubernetes/ingress-nginx/stargazers)
[![GitHub stars](https://img.shields.io/badge/contributions-welcome-orange.svg)](https://github.com/kubernetes/ingress-nginx/blob/main/CONTRIBUTING.md)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fkubernetes%2Fingress-nginx.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fkubernetes%2Fingress-nginx?ref=badge_shield)

## Overview

ingress-nginx is an Ingress controller for Kubernetes using [NGINX](https://www.nginx.org/) as a reverse proxy and load balancer.

[Learn more about Ingress on the main Kubernetes documentation site](https://kubernetes.io/docs/concepts/services-networking/ingress/).

## Get started

See the [Getting Started](https://kubernetes.github.io/ingress-nginx/deploy/) document.

## Troubleshooting

If you encounter issues, review the [troubleshooting docs](docs/troubleshooting.md), [file an issue](https://github.com/kubernetes/ingress-nginx/issues), or talk to us on the [#ingress-nginx channel](https://kubernetes.slack.com/messages/ingress-nginx) on the Kubernetes Slack server.

## Changelog

See [the list of releases](https://github.com/kubernetes/ingress-nginx/releases) to find out about feature changes.
For detailed changes for each release; please check the [Changelog.md](Changelog.md) file.
For detailed changes on the `ingress-nginx` helm chart, please check the following [CHANGELOG.md](charts/ingress-nginx/CHANGELOG.md) file.

### Support Versions table 

| Ingress-NGINX version | k8s supported version        | Alpine Version | Nginx Version |
|-----------------------|------------------------------|----------------|---------------|
| v1.1.1                | 1.23, 1.22, 1.21, 1.20, 1.19 | 3.14.2         |  1.19.9†      |
| v1.1.0                | 1.22, 1.21, 1.20, 1.19       | 3.14.2         |  1.19.9†      |
| v1.0.5                | 1.22, 1.21, 1.20, 1.19       | 3.14.2         |  1.19.9†      |
| v1.0.4                | 1.22, 1.21, 1.20, 1.19       | 3.14.2         |  1.19.9†      |
| v1.0.3                | 1.22, 1.21, 1.20, 1.19       | 3.14.2         |  1.19.9†      |
| v1.0.2                | 1.22, 1.21, 1.20, 1.19       | 3.14.2         |  1.19.9†      |
| v1.0.1                | 1.22, 1.21, 1.20, 1.19       | 3.14.2         |  1.19.9†      |
| v1.0.0                | 1.22, 1.21, 1.20, 1.19       | 3.13.5         |  1.20.1       |
| v0.50.0               | 1.21, 1.20, 1.19             | 3.14.2         |  1.19.9†      |
| v0.49.3               | 1.21, 1.20, 1.19             | 3.14.2         |  1.19.9†      |
| v0.49.2               | 1.21, 1.20, 1.19             | 3.14.2         |  1.19.9†      |
| v0.49.1               | 1.21, 1.20, 1.19             | 3.14.2         |  1.19.9†      |
| v0.49.0               | 1.21, 1.20, 1.19             | 3.13.5         |  1.20.1       |
| v0.48.1               | 1.21, 1.20, 1.19             | 3.13.5         |  1.20.1       |
| v0.47.0               | 1.21, 1.20, 1.19             | 3.13.5         |  1.20.1       |

† _This build is [patched against CVE-2021-23017](https://github.com/openresty/openresty/commit/4b5ec7edd78616f544abc194308e0cf4b788725b#diff-42ef841dc27fe0b5aa2d06bd31308bb63a59cdcddcbcddd917248349d22020a3)._

See [this article](https://kubernetes.io/blog/2021/07/26/update-with-ingress-nginx/) if you want upgrade to the stable Ingress API. 

## Get Involved

Thanks for taking the time to join our community and start contributing!

- This project adheres to the [Kubernetes Community Code of Conduct](https://git.k8s.io/community/code-of-conduct.md). By participating in this project, you agree to abide by its terms.

- **Contributing**: Contributions of all kind are welcome!
  
  - Read [`CONTRIBUTING.md`](CONTRIBUTING.md) for information about setting up your environment, the workflow that we expect, and instructions on the developer certificate of origin that we require.

  - Join our Kubernetes Slack channel for developer discussion : [#ingress-nginx-dev](https://kubernetes.slack.com/archives/C021E147ZA4).
  
  - Submit github issues for any feature enhancements, bugs or documentation problems. Please make sure to read the [Issue Reporting Checklist](https://github.com/kubernetes/ingress-nginx/blob/main/CONTRIBUTING.md#issue-reporting-guidelines) before opening an issue. Issues not conforming to the guidelines **may be closed immediately**.

- **Support**: Join the the [#ingress-nginx-users](https://kubernetes.slack.com/messages/CANQGM8BA/) channel inside the [Kubernetes Slack](http://slack.kubernetes.io/) to ask questions or get support from the maintainers and other users.
  
  - The [github issues](https://github.com/kubernetes/ingress-nginx/issues) in the repository are **exclusively** for bug reports and feature requests.

- **Discuss**: Tweet using the `#IngressNginx` hashtag.

## License

[Apache License 2.0](https://github.com/kubernetes/ingress-nginx/blob/main/LICENSE)
