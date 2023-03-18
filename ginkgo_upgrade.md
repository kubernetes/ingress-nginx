# GINKGO UPGRADE

#### Bumping ginkgo in the project requires four PRs.

## 1. Dependabot PR

- Dependabot automatically updates `ginkgo version` but only in [go.mod ](go.mod) and [go.sum ](go.sum) files.
- This is an automatically generated PR by Dependabot but it needs approval from maintainers to get merged.

## 2. Edit-hardcoded-version PR

### a. Make changes to appropriate files in required directories

- Make changes in files where gingko version is hardcoded. These files are :
    - [run-in-docker.sh ](build/run-in-docker.sh)
    - [Dockerfile ](images/test-runner/rootfs/Dockerfile)
    - [run.sh ](test/e2e/run.sh)
    - [run-chart-test.sh ](test/e2e/run-chart-test.sh)

### b. Create PR

- Open pull request(s) accordingly, to fire cloudbuild for building the component's image (if applicable).

### c. Merge

- Merging will fire cloudbuild, which will result in images being promoted to the  [staging container registry](https://console.cloud.google.com/gcr/images/k8s-staging-ingress-nginx).

### d. Make sure cloudbuild is a success

- Wait for [cloud build](https://console.cloud.google.com/cloud-build/builds?project=k8s-staging-ingress-nginx). If you don't have access to cloudbuild, you can also have a look at [this](https://prow.k8s.io/?repo=kubernetes%2Fingress-nginx&job=post-*), to see the progress of the build.

- Proceed only after cloud-build is successful in building a new Ingress-Nginx-Controller image.


## 3. PROMOTE the new-testrunner-image PR:

Promoting the images basically means that images, that were pushed to staging container registry in the steps above, now are also pushed to the public container registry. Thus are publicly available. Follow these steps to promote images:
- When you make changes to the `Dockerfile` or other core content under [images directory ](images), it generates a new image in google cloudbuild. This is because kubernetes projects need to use the infra provided for the kubernetes projects. The new image is always only pushed to the staging repository of K8S. From the staging repo, the new image needs to be promoted to the production repo. And once promoted, its possible to  use the sha of the new image in the code.

### a. Get the sha

- Get the sha of the new image(s) of the controller, from the cloudbuild, from steps above

    - The sha is available in output from [cloud build](https://console.cloud.google.com/cloud-build/builds?project=k8s-staging-ingress-nginx)

    - The sha is also visible [here](https://console.cloud.google.com/gcr/images/k8s-staging-ingress-nginx/global/e2e-test-runner)

    - The sha is also visible [here]((https://prow.k8s.io/?repo=kubernetes%2Fingress-nginx&job=post-*)), after cloud build is finished. Click on the respective job, go to `Artifacts` section in the UI, then again `artifacts` in the directory browser. In the `build.log` at the very bottom you see something like this:

  ```
  ...
  pushing manifest for gcr.io/k8s-staging-ingress-nginx/controller:v1.0.2@sha256:e15fac6e8474d77e1f017edc33d804ce72a184e3c0a30963b2a0d7f0b89f6b16
  ...
  ```

### b. Add the new image to [k8s.io](http://github.com/kubernetes/k8s.io)

- The sha(s) from the step before (and the tag(s) for the new image(s) have to be added, as a new line, in a file, of the [k8s.io](http://github.com/kubernetes/k8s.io) project of Kubernetes organization.

- Fork that other project (if you don't have a fork already).

- Other project to fork  [Github repo kubernetes/k8s.io](http://github.com/kubernetes/k8s.io)

- Fetch --all and rebase to upstream if already forked.

- Create a branch in your fork, named as the issue number for this release

- In the related branch, of your fork, edit the file k8s.gcr.io/images/k8s-staging-ingress-nginx/images.yaml.

- For making it easier, you can edit your branch directly in the browser. But be careful about making any mistake.

- Insert the sha(s) & the tag(s), in a new line, in this file [Project kubernetes/k8s.io Ingress-Nginx-Controller Images](https://github.com/kubernetes/k8s.io/blob/main/k8s.gcr.io/images/k8s-staging-ingress-nginx/images.yaml)  Look at this [example PR and the diff](https://github.com/kubernetes/k8s.io/pull/4499) to see how it was done before

- Save and commit

### c. Create PR

- Open pull request to promote the new e2e-test-runner image.

### d. Merge

- Merge success is required.

- Proceed only after cloud-build is successful in building a new e2e-test-runner image.


## 4. Change testrunner-image-sha PR

### a. Get the sha

- Make sure to get the tag and sha of the promoted image from the step before, either from cloudbuild or from [here](https://console.cloud.google.com/gcr/images/k8s-artifacts-prod/us/ingress-nginx/e2e-test-runner).

### a. Make sure your git workspace is ready

- Get your git workspace ready

    - If not using a pre-existing fork, then Fork the repo kubernetes/ingress-nginx

    - Clone (to laptop or wherever)

    - Add upstream

    - Set upstream url to no_push

    - Checkout & switch to branch, named as per related new-release-issue-number

    - If already forked, and upstream already added, then `git fetch --all` and `git rebase upstream/main` (not  origin)

    - Checkout a branch in your fork's clone

    - Perform any other diligence as needed

- Prefer to edit only and only in your branch, in your Fork

### b. Change testrunner-image-sha

- You need update the testrunner-image-sha in the following files :

    - [run-in-docker.sh](https://github.com/kubernetes/ingress-nginx/blob/main/build/run-in-docker.sh#L41)
    - [Makefile](https://github.com/kubernetes/ingress-nginx/blob/main/test/e2e-image/Makefile#L3)

### c. Create PR

- Look at this PR for how it was done before [example PR](https://github.com/kubernetes/ingress-nginx/pull/9444)
- Create a PR

### d. Merge

- Finally merge the PR.

## END ##