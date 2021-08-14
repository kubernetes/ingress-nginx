## 1. [NGINX Base Image](https://github.com/kubernetes/ingress-nginx/tree/main/images/nginx)

A. Open pull request
  
  If you are updating any component in [build.sh](images/nginx/rootfs/build.sh) please also update the SHA256 checksum of that component as well, the cloud build will fail with an exit 10 if not.  
Example [NGINX_VERSION](images/nginx/rootfs/build.sh#L21),
[SHA256](images/nginx/rootfs/build.sh#L124) 

B. Merge

C. Wait for [cloud build](https://console.cloud.google.com/cloud-build/builds?project=k8s-staging-ingress-nginx)

D. Promote Images:

  Open pull request to promote staging image:
  [add sha - version](https://github.com/kubernetes/k8s.io/blob/main/k8s.gcr.io/images/k8s-staging-ingress-nginx/images.yaml#L1)

The sha is available in output from [cloud build](https://console.cloud.google.com/cloud-build/builds?project=k8s-staging-ingress-nginx)

## 2. Change to other Images:

   
A. [e2e](https://github.com/kubernetes/ingress-nginx/tree/main/test/e2e-image)

B. [test-runner](https://github.com/kubernetes/ingress-nginx/tree/main/images/test-runner)
C. [echo](https://github.com/kubernetes/ingress-nginx/tree/main/images/echo)

D. [cfssl](https://github.com/kubernetes/ingress-nginx/tree/main/images/cfssl)

E. [fastcgi-helloserver](https://github.com/kubernetes/ingress-nginx/tree/main/images/fastcgi-helloserver)

F. [httpbin](https://github.com/kubernetes/ingress-nginx/tree/main/images/httpbin)

G. Open pull request

H. Merge

I. Wait for [cloud build](https://console.cloud.google.com/cloud-build/builds?project=k8s-staging-ingress-nginx)

J. Promote Images:

- Open pull request to promote [staging image](https://github.com/kubernetes/k8s.io/blob/main/k8s.gcr.io/images/k8s-staging-ingress-nginx/images.yaml)
    * e2e-test-runner
    * e2e-test-cfssl
    * e2e-test-echo
    * e2e-test-fastcgi-helloserver
    * e2e-test-httpbin

## 3. Update references to e2e-test-runner image:

* [e2e-image](https://github.com/kubernetes/ingress-nginx/blob/main/test/e2e-image/Dockerfile#L1)
* [run-in-docker.sh](https://github.com/kubernetes/ingress-nginx/blob/main/build/run-in-docker.sh#L37)

## 4. Promote Controller Image:

- Get the sha of the new image of the controller from the cloudbuild, from
    steps above

- This sha (and the tag for the new image) has to be inserted, as a new line, in a file, in another project of
    Kubernetes [Ingress-Nginx-Controller Images](https://github.com/kubernetes/k8s.io/blob/main/k8s.gcr.io/images/k8s-staging-ingress-nginx/images.yaml)

- Fork the repo [Github repo kubernetes/k8s.io](http://github.com/kubernetes/k8s.io)

- In your fork, edit the file /k8s.gcr.io/images/k8s-staging-ingress-nginx/images.yaml. Insert the sha & the tag, in a new line. Look at this [example PR and the diff](https://github.com/kubernetes/k8s.io/pull/2536) to see how it was done before 

- Open pull request to promote the new controller image

- Merge

## 5. Prepare for a new Release

- This involves editing of several different files so carefully follow the steps below and double check all changes with diff/grep etc., repeatedly.

- Get your workspace ready to make edits

  - If not using a pre-existing fork, then Fork the repo kubernetes/ingress-nginx

  - Clone (to laptop or wherever)

  - Add upstream

  - Set upstream url to no_push

  - Checkout & switch to branch named as per related new-release-issue-number

  - If already forked, and upstream already added, then `git fetch --all` and `git rebase upstream` (not  origin)

  - Checkout a branch in your fork's clone

- Prefer to edit only and only in your branch, in your Fork

### A. Edit the semver tag
  - [TAG](https://github.com/kubernetes/ingress-nginx/blob/main/TAG#L1)

### B. Edit the helm Chart
  - Change the below mentioned [Fields in Chart.yaml](https://github.com/kubernetes/ingress-nginx/blob/main/charts/ingress-nginx/Chart.yaml)
    - version
    - appVersion
    - kubeVersion (**ONLY if applicable**)
    - annotations
      - artifacthub.io/prerelease: "true" 
      - artifacthub.io/changes: |
        - Add the PRs merged after previous release

### C. Edit the values.yaml
  - [Fields to edit in values.yaml](https://github.com/kubernetes/ingress-nginx/blob/main/charts/ingress-nginx/values.yaml)

    - tag

    - digest

### D. Edit the static manifests

  - Prepare to use a script to update the edit the static manifests and set the "image", "digest", "version" etc. fields to the desired value.


  - This script depends on python and a specific python package `pip3 install ruamel.yaml`

  - Execute the script to update static manifests using that script [generate-deploy-scripts.sh](https://github.com/kubernetes/ingress-nginx/blob/main/hack/generate-deploy-scripts.sh)
  - Open some of the manifests and check if the script worked properly

  - Use grep -ir to search for any misses by the script or undesired changes

  - The script should properly set the image and the digest fields to the desired tag and semver

  - Manually fix one problem that the script can not take care of.

    - This problem is wrong formatting of a snippet in the file [deploy-tls-termination.yaml](https://github.com/kubernetes/ingress-nginx/blob/main/deploy/static/provider/aws/deploy-tls-termination.yaml)
    - In the configMap section, for the configMap named ingress-nginx-controller, the "configMap.data" spec has a snippet

    - This snippet becomes a single line, formatted with the newline character "\n"

    - That single line formatted with "\n" needs to be changed as it does not meet yaml requirements

    - At the time of writing this doc, the 'configMap.data' spec is at line number 39.

    - So editing begins at line 40 (at the time of writing this doc)

    - Make that snippet look like this ;
      ```
      data:
        http-snippet:|
          server{
          listen 2443;
          return 308 https://$host$request_uri;
          }
      ```

### E. Edit the changelog
  [Changelog.md](https://github.com/kubernetes/ingress-nginx/blob/main/Changelog.md) 
    - Add the PRs merged after previous release

### F. Edit the Documentation:
    - Update the version in [docs/deploy/index.md](docs/deploy/index.md)
    - Update Supported versions in the Support Versions table in the README.md 

### G. Edit the stable.txt file(if applicable), in the root of the repo, to reflect the release to be created

### H. Edit the ghpages for Github release (if applicable)

### I. Open PR for releasing the Ingress-Nginx-Controller ;
  - Look at this PR for how it was done before [example PR](https://github.com/kubernetes/ingress-nginx/pull/7490)
  - Create a PR 
  - Merge
