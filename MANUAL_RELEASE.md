# RELEASE PROCESS

## 1. BUILD the new Ingress-Nginx-Controller image

### a. Make changes in codebase

- Make changes as per issue

### b. Make changes to appropriate files in [images directory ](images)

- Make changes in /images

### c. Create Pull Request

- Open a Pull Request for your changes considering the following steps to fire cloudbuild of a new image for the Ingress-Nginx-Controller:

  - In case of rare CVE fix or other reason to rebuild the nginx-base-image itself, look at the /images directory  [NGINX Base Image](https://github.com/kubernetes/ingress-nginx/tree/main/images/nginx).

  - Example [NGINX_VERSION](images/nginx/rootfs/build.sh#L21), [SHA256](images/nginx/rootfs/build.sh#L124).

  - If you are updating any component in [build.sh](images/nginx/rootfs/build.sh) please also update the SHA256 checksum of that component as well, the cloud build will fail with an exit 10 if not.

### d. Merge

- Merging will fire cloudbuild, which will result in images being promoted to the [staging container registry](https://console.cloud.google.com/gcr/images/k8s-staging-ingress-nginx).

### e. Make sure cloudbuild is a success

- Wait for [cloud build](https://console.cloud.google.com/cloud-build/builds?project=k8s-staging-ingress-nginx). If you don't have access to cloudbuild, you can also have a look at [this](https://prow.k8s.io/?repo=kubernetes%2Fingress-nginx&job=post-*), to see the progress of the build.

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

- Merging will fire cloudbuild, which will result in images being promoted to the [staging container registry](https://console.cloud.google.com/gcr/images/k8s-staging-ingress-nginx).

### e. Make sure cloudbuild is a success

- Wait for [cloud build](https://console.cloud.google.com/cloud-build/builds?project=k8s-staging-ingress-nginx). If you don't have access to cloudbuild, you can also have a look at [this](https://prow.k8s.io/?repo=kubernetes%2Fingress-nginx&job=post-*), to see the progress of the build.

- Proceed only after cloud-build is successful in building a new Ingress-Nginx-Controller image.


## 3. PROMOTE the Image(s):

Promoting the images basically means that images, that were pushed to staging container registry in the steps above, now are also pushed to the public container registry. Thus are publicly available. Follow these steps to promote images:

### a. Get the sha

- Get the sha of the new image(s) of the controller, (and any other component image IF APPLICABLE to release), from the cloudbuild, from steps above

  - The sha is available in output from [cloud build](https://console.cloud.google.com/cloud-build/builds?project=k8s-staging-ingress-nginx)

  - The sha is also visible here https://console.cloud.google.com/gcr/images/k8s-staging-ingress-nginx/global/controller

  - The sha is also visible [here](https://prow.k8s.io/?repo=kubernetes%2Fingress-nginx&job=post-*), after cloud build is finished. Click on the respective job, go to `Artifacts` section in the UI, then again `artifacts` in the directory browser. In the `build.log` at the very bottom you see something like this:

  ```
  ...
  pushing manifest for gcr.io/k8s-staging-ingress-nginx/controller:v1.0.2@sha256:e15fac6e8474d77e1f017edc33d804ce72a184e3c0a30963b2a0d7f0b89f6b16
  ...
  ```

### b. Add the new image to [k8s.io](http://github.com/kubernetes/k8s.io)

- The sha(s) from the step before (and the tag(s) for the new image(s) have to be added, as a new line, in a file, of the [k8s.io](http://github.com/kubernetes/k8s.io) project of Kubernetes organization.

- Fork that other project (if you don't have a fork already).

- Other project to fork  [GitHub repo kubernetes/k8s.io](http://github.com/kubernetes/k8s.io)

- Fetch --all and rebase to upstream if already forked.

- Create a branch in your fork, named as the issue number for this release

- In the related branch, of your fork, edit the file /registry.k8s.io/images/k8s-staging-ingress-nginx/images.yaml.

- For making, it easier, you can edit your branch directly in the browser. But be careful about making any mistake.

- Insert the sha(s) & the tag(s), in a new line, in this file [Project kubernetes/k8s.io Ingress-Nginx-Controller Images](https://github.com/kubernetes/k8s.io/blob/main/k8s.gcr.io/images/k8s-staging-ingress-nginx/images.yaml)  Look at this [example PR and the diff](https://github.com/kubernetes/k8s.io/pull/2536) to see how it was done before

- Save and commit

### c. Create PR

- Open pull request to promote the new controller image.

### d. Merge

- Merge success is required for next step

- Proceed only after cloud-build is successful in building a new Ingress-Nginx-Controller image.


## 4. PREPARE for a new Release

- Make sure to get the tag and sha of the promoted image from the step before, either from cloudbuild or from [here](https://console.cloud.google.com/gcr/images/k8s-artifacts-prod/us/ingress-nginx/controller).

- This involves editing of several files. So carefully follow the steps below and double check all changes with diff/grep etc., repeatedly. Mistakes here impact endusers.

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
  - Change the below-mentioned [Fields in Chart.yaml](https://github.com/kubernetes/ingress-nginx/blob/main/charts/ingress-nginx/Chart.yaml)
    - version
    - appVersion
    - kubeVersion (**ONLY if applicable**)
    - annotations
      - artifacthub.io/prerelease: "true"
      - artifacthub.io/changes: |
        - Replace this line and other lines under this annotation with the Changelog. One process to generate the Changelog is described below
          - Install and configure GitHub cli as per the docs of gh-cli  https://cli.github.com/,
          - Change dir to your clone, of your fork, of the ingress-nginx project
          - Run the below command and save the output to a txt file

            ```
            gh pr list -R kubernetes/ingress-nginx -s merged -L 38 -B main | cut -f1,2 | tee ~/Downloads/prlist.txt
            ```
            - The -L 38 was used for 2 reasons.
              - Default number of results is 30 and there were more than 30 PRs merged while releasing v1.1.1. If you see the current/soon-to-be-old changelog, you can look at the most recent PR number that has been accounted for already, and start from after that last accounted for PR.
              -  The other reason to use -L 38 was to ommit the 39th, the 40th and the 41st line in the resulting list. These were non-relevant PRs.
            - If you save the output of above command to a file called prlist.txt. It looks somewhat like this ;

              ```
              % cat ~/Downloads/prlist.txt 
              8129    fix syntax in docs for multi-tls example
              8120    Update go in runner and release v1.1.1
              8119    Update to go v1.17.6
              8118    Remove deprecated libraries, update other libs
              8117    Fix codegen errors
              8115    chart/ghaction: set the correct permission to have access to push a release 
              ....
              ```
              You can delete the lines, that refer to PRs of the release process itself. We only need to list the feature/bugfix PRs. You can also delete the lines that are housekeeping or not really worth mentioning in the changelog.
          -  you use some easy automation in bash/python/other, to get the PR-List that can be used in the changelog. For example, its possible to use a bash scripty way, seen below, to convert those plaintext PR numbers into clickable links. 

            ```
            #!/usr/bin/bash

            file="$1"

            while read -r line; do
              pr_num=`echo "$line" | cut -f1`
              pr_title=`echo "$line" | cut -f2`
              echo "[$pr_num](https://github.com/kubernetes/ingress-nginx/pull/$pr_num) $pr_title"
            done <$file

            ```
          - There was a parsing issue and path issue on MacOS, so above scrpt had to be modified and MacOS monterey compatible script is below ;

            ```
            #!/bin/bash

            file="$1"

            while read -r line; do
              pr_num=`echo "$line" | cut -f1`
              pr_title=`echo "$line" | cut -f2`
              echo \""[$pr_num](https://github.com/kubernetes/ingress-nginx/pull/$pr_num) $pr_title"\"
            done <$file

            ```
          - If you saved the bash script content above, in a file like `$HOME/bin/prlist_to_changelog.sh`, then you could execute a command like this to get your prlist in a text file called changelog_content.txt;`

          ```
          prlist_to_changelog.sh ~/Downloads/prlist.txt | tee ~/Downloads//changelog_content.txt
          ```

### d. Edit the values.yaml and run helm-docs
  - [Fields to edit in values.yaml](https://github.com/kubernetes/ingress-nginx/blob/main/charts/ingress-nginx/values.yaml)

    - tag
    - digest

  - [helm-docs](https://github.com/norwoodj/helm-docs) is a tool that generates the README.md for a helm-chart automatically. In the CI pipeline workflow of github actions (/.github/workflows/ci.yaml), you can see how helm-docs is used. But the CI pipeline is not designed to make commits back into the project. So we need to run helm-docs manually, and check in the resulting autogenerated README.md at the path /charts/ingress-nginx/README.md 
    ```
          GOBIN=$PWD GO111MODULE=on go install github.com/norwoodj/helm-docs/cmd/helm-docs@v1.11.0
          ./helm-docs --chart-search-root=${GITHUB_WORKSPACE}/charts
          git diff --exit-code
          rm -f ./helm-docs
    ```
    Watchout for mistakes like leaving the helm-docs executable in your clone workspace or not checking the new README.md manually etc.

### e. Edit the static manifests

  - Prepare to use a script to update the edit the static manifests and set the "image", "digest", "version" etc. fields to the desired value.

  - This script depends on kustomize and helm. The versions are pinned in `hack/.tool-versions` and you can use [asdf](https://github.com/asdf-vm/asdf#asdf) to install them

  - Execute the script to update static manifests using that script [hack/generate-deploy-scripts.sh](https://github.com/kubernetes/ingress-nginx/blob/main/hack/generate-deploy-scripts.sh)
  - Open some of the manifests and check if the script worked properly

  - Use `grep -ir image: | less` on the deploy directory, to view for any misses by the script on image digest value or other undesired changes. The script should properly set the image and the digest fields to the desired tag and semver


### f. Edit the changelog

  [Changelog.md](https://github.com/kubernetes/ingress-nginx/blob/main/Changelog.md)
- Each time a release is made, a new section is added to the Changelog.md file
- A new section in the Changelog.md file consists of 3 components listed below
  - the "Image"
  - the "Description"
  - the "PRs list"
- Look at the previous content to understand what the 3 components look like.
- You can easily get the "Image" from a yaml manifest but be sure to look at a manifest in your git clone now and not the upstream on github. This is because, if you are following this documentation, then you generated manifests with new updated digest for the image, in step 4e above. You also most likely promoted the new image in a step above. Look at the previous release section in Changelog.md. The format looks like `registry.k8s.io/ingress-nginx/controller:.......`. One example of a yaml file to look at is /deploy/static/provider/baremetal/deploy.yaml (in your git clone branch and not on the upstream).
- Next, you need to have a good overview of the changes introduced in this release and based on that you write a description. Look at previous descriptions. Ask the ingress-nginx-dev channel if required.
- And then you need to add a list of the PRs merged, since the previous release.
- One process to generate this list of PRs is already described above in step 4c. So if you are following this document, then you have done this already and very likely have retained the file containing the list of PRs, in the format that is needed.

### g. Edit the Documentation:

- Update the version in [docs/deploy/index.md](docs/deploy/index.md)
- Update Supported versions in the Support Versions table in the README.md
- Execute the script to update e2e docs [hack/generate-e2e-suite-doc.sh](https://github.com/kubernetes/ingress-nginx/blob/main/hack/generate-e2e-suite-doc.sh)

### h. Update README.md

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

## TODO
- Automate & simplify as much as possible, whenever possible, however possible
