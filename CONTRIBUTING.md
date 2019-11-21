# Contributing Guidelines

Read the following guide if you're interested in contributing to Ingress. [Make Ingress-Nginx Work for you, and the Community](https://youtu.be/GDm-7BlmPPg) from KubeCon Europe 2018 is a great video to get you started!!

## Contributor License Agreements

We'd love to accept your patches! Before we can take them, we have to jump a couple of legal hurdles.

Please fill out either the individual or corporate Contributor License Agreement (CLA).

  * If you are an individual writing original source code and you're sure you own the intellectual property, then you'll need to sign an [individual CLA](https://identity.linuxfoundation.org/projects/cncf).
  * If you work for a company that wants to allow you to contribute your work, then you'll need to sign a [corporate CLA](https://identity.linuxfoundation.org/node/285/organization-signup).

Follow either of the two links above to access the appropriate CLA and instructions for how to sign and return it. Once we receive it, we'll be able to accept your pull requests.

***NOTE***: Only original source code from you and other people that have signed the CLA can be accepted into the main repository.

## Finding Things That Need Help

If you're new to the project and want to help, but don't know where to start, we have a semi-curated list of issues that should not need deep knowledge of the system. [Have a look and see if anything sounds interesting](https://github.com/kubernetes/ingress-nginx/issues?utf8=%E2%9C%93&q=is%3Aopen%20is%3Aissue%20label%3A%22help+wanted%22). Alternatively, read some of the docs on other controllers and try to write your own, file and fix any/all issues that come up, including gaps in documentation!

## Contributing a Patch

1. If you haven't already done so, sign a Contributor License Agreement (see details above).
1. Read the [Ingress development guide](docs/development.md).
1. Fork the desired repo, develop and test your code changes.
1. Submit a pull request.

All changes must be code reviewed. Coding conventions and standards are explained in the official [developer docs](https://github.com/kubernetes/community/tree/master/contributors/devel). Expect reviewers to request that you avoid common [go style mistakes](https://github.com/golang/go/wiki/CodeReviewComments) in your PRs.

### Merge Approval

Ingress collaborators may add "LGTM" (Looks Good To Me) or an equivalent comment to indicate that a PR is acceptable. Any change requires at least one LGTM.  No pull requests can be merged until at least one Ingress collaborator signs off with an LGTM.

## Support Channels

Whether you are a user or contributor, official support channels include:

- GitHub issues: https://github.com/kubernetes/ingress-nginx/issues/new
- Slack: kubernetes-users room in the [Kubernetes Slack](http://slack.kubernetes.io/)
- Post: [Kubernetes Forum](https://discuss.kubernetes.io)

Before opening a new issue or submitting a new pull request, it's helpful to search the project - it's likely that another user has already reported the issue you're facing, or it's a known issue that we're already aware of.
