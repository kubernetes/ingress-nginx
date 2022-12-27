# GINKGO UPGRADE

## Bumping ginkgo in the project requires 4 PRs.

### a. Dependabot PR   

- Dependabot automatically updates `ginkgo version` but only in [go.mod ](go.mod) and [go.sum ](go.sum) files.

### b. Edit-hardcoded-version PR

- After that you need to make version changes in files where gingko version is hardcoded. These files are [build/run-in-docker.sh ](run-in-docker.sh), [images/test-runner/rootfs/Dockerfile ](Dockerfile), [test/e2e/run.sh ](run.sh) and [test/e2e/run-chart-test.sh ](run-chart-test.sh)

### c. Promote-new-testrunner-image PR 

- When you make changes to the `Dockerfile` or other core content under [images directory ](images), it generates a new image in google cloudbuild. This is because kubernetes projects need to use the infra provided for the kubernetes projects. The new image is always only pushed to the staging repository of K8S. From the staging repo, the new image needs to be promoted to the production repo. And once promoted, its possible to  use the sha of the new image in the code.

### d. Change-testrunner-image-sha PR

- In this case look for the sha of the test-runner image. Although new image is built, it needs to be promoted and then the sha in the test code needs to change to new image's sha.  Only then ginkgo updated version from the new test-runner becomes available for use during e2e tests
- Look at the RELEASE.md for the link to the GCP staging repo to check sha of new image

