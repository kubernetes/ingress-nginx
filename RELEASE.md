# RELEASE PROCESS

## 1. BUILD the new Ingress-Nginx-Controller image 

### a. Make changes in codebase

- Make changes as per issue

### b. Make changes to appropriate files in [images directory ](images)

- Make changes in /images 

### c. Create PR

- Open a PR to fire cloudbuild of a new image for the Ingress-Nginx-Controller

  - In case of rare CVE fix or other reason to rebuild the nginx-base-image itself, look at the /images directory  [NGINX Base Image](https://github.com/kubernetes/ingress-nginx/tree/main/images/nginx)

  - Example [NGINX_VERSION](images/nginx/rootfs/build.sh#L21),

  - [SHA256](images/nginx/rootfs/build.sh#L124) 

  - If you are updating any component in [build.sh](images/nginx/rootfs/build.sh) please also update the SHA256 checksum of that component as well, the cloud build will fail with an exit 10 if not.  

### d. Merge

- Merging success should fire cloudbuild

### e. Make sure cloudbuild is a success

- Wait for [cloud build](https://console.cloud.google.com/cloud-build/builds?project=k8s-staging-ingress-nginx)

- Proceed only after cloud-build is successful in building a new Ingress-Nginx-Controller image.


## 2. If applicable, BUILD other images

- If applicable, then build a new image of any other related component, ONLY IF APPLICABLE TO THE RELEASE

### a. If applicable then make changes in relevant codebase

- Change code as per issue

### b. Make changes to appropriate files in [images directory ](images)

- Sometimes, you may also be needing to rebuild, images for one or multiple other related components of the Ingress-Nginx-Controller ecosystem. Make changes to the required files in the /images directory, if/as applicable, in the context of the release you are attempting.  :
   
  - [e2e](https://github.com/kubernetes/ingress-nginx/tree/main/test/e2e-image)

    - Update references to e2e-test-runner image [If applicable] :

      - [e2e-image](https://github.com/kubernetes/ingress-nginx/blob/main/test/e2e-image/Dockerfile#L1)
      - [run-in-docker.sh](https://github.com/kubernetes/ingress-nginx/blob/main/build/run-in-docker.sh#L37)

  - [test-runner](https://github.com/kubernetes/ingress-nginx/tree/main/images/test-runner)

  - [echo](https://github.com/kubernetes/ingress-nginx/tree/main/images/echo)

  - [cfssl](https://github.com/kubernetes/ingress-nginx/tree/main/images/cfssl)

  - [fastcgi-helloserver](https://github.com/kubernetes/ingress-nginx/tree/main/images/fastcgi-helloserver)

  - [httpbin](https://github.com/kubernetes/ingress-nginx/tree/main/images/httpbin)

  - [kube-webhook-certgen](https://github.com/kubernetes/ingress-nginx/tree/main/images/kube-webhook-certgen)

### c. Create PR

- Open pull request(s) accordingly, to fire cloudbuild for rebuilding the component's image (if applicable).

### d. Merge

- Merging success should fire cloudbuild

### e. Make sure cloudbuild is a success

- Wait for [cloud build](https://console.cloud.google.com/cloud-build/builds?project=k8s-staging-ingress-nginx)

- Proceed only after cloud-build is successful in building a new Ingress-Nginx-Controller image.


## 3. PROMOTE the Image(s):

### a. Get the sha

- Get the sha of the new image(s) of the controller, (and any other component image IF APPLICABLE to release), from the cloudbuild, from steps above 

  - The sha is available in output from [cloud build](https://console.cloud.google.com/cloud-build/builds?project=k8s-staging-ingress-nginx)

  - The sha is also visible here https://console.cloud.google.com/gcr/images/k8s-staging-ingress-nginx/global/controller

### b. Insert the sha(s) in another project

- This sha(s) (and the tag(s) for the new image(s) has to be inserted, as a new line, in a file, in another project of Kubernetes. 

- Fork that other project (if you don't have a fork already).

- Other project to fork  [Github repo kubernetes/k8s.io](http://github.com/kubernetes/k8s.io)

- Fetch --all and rebase to upstream if already forked.

- Create a branch in your fork, named as the issue number for this release

- In the related branch, of your fork, edit the file /k8s.gcr.io/images/k8s-staging-ingress-nginx/images.yaml.

- For making it easier, you can edit your branch directly in the browser. But be careful about making any mistake.

- Insert the sha(s) & the tag(s), in a new line, in this file [Project kubernetes/k8s.io Ingress-Nginx-Controller Images](https://github.com/kubernetes/k8s.io/blob/main/k8s.gcr.io/images/k8s-staging-ingress-nginx/images.yaml)  Look at this [example PR and the diff](https://github.com/kubernetes/k8s.io/pull/2536) to see how it was done before 

- Save and commit

### c. Create PR

- Open pull request to promote the new controller image.

### d. Merge

- Merge success is required for next step

- Proceed only after cloud-build is successful in building a new Ingress-Nginx-Controller image.


## 4. PREPARE for a new Release

- This involves editing of several different files. So carefully follow the steps below and double check all changes with diff/grep etc., repeatedly. Mistakes here impact endusers.

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

### b. Edit the semver tag
  - [TAG](https://github.com/kubernetes/ingress-nginx/blob/main/TAG#L1)

### c. Edit the helm Chart
  - Change the below mentioned [Fields in Chart.yaml](https://github.com/kubernetes/ingress-nginx/blob/main/charts/ingress-nginx/Chart.yaml)
    - version
    - appVersion
    - kubeVersion (**ONLY if applicable**)
    - annotations
      - artifacthub.io/prerelease: "true" 
      - artifacthub.io/changes: |
        - Add the titles of the PRs merged after previous release

### d. Edit the values.yaml
  - [Fields to edit in values.yaml](https://github.com/kubernetes/ingress-nginx/blob/main/charts/ingress-nginx/values.yaml)

    - tag
    - digest

### e. Edit the static manifests

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

### f. Edit the changelog
  [Changelog.md](https://github.com/kubernetes/ingress-nginx/blob/main/Changelog.md) 
    - Add the PRs merged after previous release
    - One useful command to get this list is
      ```
      git log controller-v0.48.1..HEAD --pretty=%s
      ```

### g. Edit the Documentation:
    - Update the version in [docs/deploy/index.md](docs/deploy/index.md)
    - Update Supported versions in the Support Versions table in the README.md 

### h. Edit stable.txt

- Edit the [stable.txt](stable.txt) file(if applicable), in the root of the repo, to reflect the release to be created
- Criteria is a release that has been GA for a while but reported issues are not bugs but mostly /kind support or feature

### i. Update README.md
- Update the table in README.md in the root of the projet to reflect the support matrix. Add the new release version and details in there.

## 5. RELEASE new version

### a. Create PR

- Open PR for releasing the new version of the Ingress-Nginx-Controller ;
  - Look at this PR for how it was done before [example PR](https://github.com/kubernetes/ingress-nginx/pull/7490)
  - Create a PR 

### b. Merge

- Merge should produce manifests as well as chart
- Check
 - `helm repo update`
 - `helm search repo ingress-nginx`

## 6. Github release

- Release to github

- Edit the ghpages file as needed
