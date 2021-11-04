# Triage Process

As any kind of contributor (triage, reviewer ...), always have in mind that if a user came to us and raised an issue, the user may have a real problem. We must assume that, and not the opposite (the user needs to prove to us that this is a bug). Keeping that in mind, **be nice with users, even if you don’t agree with them**

Note that this guide refers to contributing through issue triaging. If you are interested in contributing to actual sources of the repository, see [this guide](./CONTRIBUTING.md). 

## General Information 

The triage process of the ingress-nginx maintainers is based on the [triage process guidelines](https://github.com/kubernetes/community/blob/master/contributors/guide/issue-triage.md) of the Kubernetes community  

However the exact process of the ingress-nginx maintainers may differ in certain aspects. This doc gives a more precise overview on how the ingress-nginx maintainers approach the issue triage process and other processes that are related. 

## Triage Flow (Issues)

This section describes the different stages of the triage flow for issues.

### Prepare Issues
New issues come in with the labels `needs-triage` and `needs-priority` and one of: `kind/bug`, `kind/feature` or `kind/support`. Unfortunately there are also some legacy issues that only have a `kind/*` label but neither `needs-triage` nor `needs-priority` . However for every issue that does not have the `triage-accepted` label the following steps have to be done to prepare them for further processing:  

* Filter for issues [without the `triage-accepted`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+-label%3Atriage%2Faccepted+is%3Aissue) label.
* Check if all necessary information are available. This is basically true, if people filled out the issue template correctly. If necessary information is missing, ask the author to add the missing information and add the label `triage/needs-information` if not already present. If already present, send the author a friendly reminder to add those. 
* Check if the used versions of ingress-nginx and Kubernetes is supported. Note that [we only support n-3 versions](https://github.com/kubernetes/ingress-nginx#support-versions-table). If the version is not supported, ask the author to upgrade to newer versions and see if the error still persists. 
* Read through the issue description and comments briefly to understand what the issue is about. Also check if the kind and area is correct, and adjust it if necessary. If the issue is understandable add the label `triage-accepted`.
* If at any point you don't know how to proceed with an issue during the triage process, tag one of the [core maintainers](OWNERS_ALIASES) in the issue to raise attention or alternatively come to [this slack channel](https://kubernetes.slack.com/archives/C021E147ZA4) which may be the quicker way as people tend to miss github notifications. 

Note: Issues that are stale for 90 days are being closed automatically. However we could be missing a bug here, so from time to time it makes sense to go over the closed ones and see if there is something important. Use [this filter](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aclosed+is%3Aissue+label%3Alifecycle%2Frotten+) to find those. 

Who and When?
* Basically everyone who wants to contribute can do the mentioned steps at any time.

### Issue Prioritization
For all issues, where all necessary information is available thus triage is accepted, we need to do some prioritization: 

* Go through all issues with label [`triage-accepted`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+is%3Aissue+label%3Atriage%2Faccepted+).  
* Add appropriate priority label: `priority/backlog`, `priority/critical-urgent`, `priority/awaiting-more-evidence`, `priority/important-longterm`, `priority/important-soon` or `good first issue`

Who and When? 
* Basically every contributor should be able to do that.
* Tricky/important ones could be brought up during community meetings

## Triage Flow (Pull Requests)

This section describes the different stages of the triage flow for pull requests.

### Prepare Pull Requests
Pull requests come in with the labels `needs-triage`, `needs-priority` and `needs-kind` and one that indicates the size(`size/*`). Unfortunately there are also some legacy pull requests that only have a `size/*` label but neither `needs-triage` nor `needs-priority` . However for every pull request that does not have the `triage-accepted` label the following steps should be done to prepare them for further processing:   

* Filter for pull requests [without the `triage-accepted`](https://github.com/kubernetes/ingress-nginx/pulls?q=is%3Aopen+-label%3Atriage%2Faccepted+is%3Apr) label.
* Check if the cla is signed and all necessary information are available. This is basically true, if people filled out the pull request template correctly. If everything is fine add the `triage-accepted` label.
* If at any point you don't know how to proceed with an issue during the triage process, tag one of the [core maintainers](OWNERS_ALIASES) in the issue to raise attention or alternatively come to [this slack channel](https://kubernetes.slack.com/archives/C021E147ZA4) which may be the quicker way as people tend to miss github notifications. 

Who and When?
* Basically everyone who wants to contribute can do the mentioned steps at any time.

### Pull Request Prioritization
For all pull requests, where all necessary information is available and cla is signed thus triage is accepted, we need to do some prioritization: 

* Go through all pull requests with label [`triage-accepted`](https://github.com/kubernetes/ingress-nginx/pulls?q=is%3Aopen+is%3Apr+label%3Atriage%2Faccepted).  
* Sync the `kind/*` and `priority/*` label from the linked issue for the pull request. If the pull request does not have any issue associated (which normally should not be the case), add an appropriate priority and kind label (one of: `priority/backlog`, `priority/critical-urgent`, `priority/important-longterm`, `priority/important-soon`) 

Who and When? 
* Basically every contributor should be able to do that.
* Tricky/important ones could be brought up during community meetings

## Labels 
Labels are helpful for issues or pull requests to indicate in which lifecycle state they are currently and to categorize them. This section describes the most important ones with the additional info about how to add those. A complete label list of the Kubernetes community can be found [here](https://github.com/kubernetes/kubernetes/labels) while a complete label list for this project can be found [here](https://github.com/kubernetes/ingress-nginx/labels). However, here the most important ones: 

* Triage:
  * [`needs-triage`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+is%3Aissue+label%3Aneeds-triage): Indicates that the issue or pull request needs triage. Automatically added. 
  * [`triage/accepted`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+label%3Atriage%2Faccepted+is%3Aissue+): Indicates that the issue is ready for further processing. Add with `/triage accepted`.
  * [`triage/needs-information`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+label%3Atriage%2Fneeds-information+is%3Aissue+): Indicates that the issue lacks information. Add with `/triage needs-information`.
* Kind: 
  * [`kind/bug`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+label%3Akind%2Fbug+is%3Aissue): Indicates that the issue is assumed to be a bug. Add with `/kind bug`. Remove with `/remove-kind bug`.
  * [`kind/feature`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+label%3Akind%2Ffeature+is%3Aissue+): Indicates that the issue is a feature request. Add with `/kind feature`. Remove with `/remove-kind feature`.
  * [`kind/documentation`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+label%3Akind%2Fdocumentation+is%3Aissue+): Indicates that the issue is documentation related. Add with `/kind documentation`. Remove with `/remove-kind documentation`.
  * [`kind/support`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+label%3Akind%2Fsupport+is%3Aissue+): Indicates the the issue is a support request. Add with `/kind support`. Remove with `/remove-kind support`.
* Area: 
  * [`area/helm`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+label%3Aarea%2Fhelm+is%3Aissue+): Indicates that the issue is related to helm charts. Add with `/area helm`.
  * [`area/lua`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+label%3Aarea%2Flua+is%3Aissue+): Indicates that the issue is related to lua. Add with `/area lua`.
  * [`area/docs`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+label%3Aarea%2Fdocs+is%3Aissue): Indicates that the issue is related to documentation. Add with `/area docs` .
* Priority: 
  * [`needs-priority`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+is%3Aissue+label%3Aneeds-priority): Indicates that the issue has no prioritization yet. Automatically added.
  * [`priority/critical-urgent`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+label%3Apriority%2Fcritical-urgent+is%3Aissue+): indicates that the issue has highest priority. Add with `/priority critical-urgent`.
  * [`priority/important-soon`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+label%3Apriority%2Fimportant-soon+is%3Aissue+): indicates that the issue should be worked on either currently soon, ideally in time for the next release. Add with `/priority important-soon`.
  * [`priority/important-longterm`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+label%3Apriority%2Fimportant-longterm+is%3Aissue+): indicates that the issue is not important for now, but should be worked on in one of the upcoming releases. Add with `/priority important-longterm`.
  * [`priority/backlog`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+label%3Apriority%2Fbacklog+is%3Aissue+): Indicates that the issue has the lowest priority. Add with `/priority backlog`.
* Other: 
  * [`help wanted`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22): indicates that the issue needs help from a contributor. Add with `/help`.
  * [`good first issue`](https://github.com/kubernetes/ingress-nginx/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22): indicates that the issue needs help from a contributor and is a good first issue for new contributors. Add with `/good-first-issue`.