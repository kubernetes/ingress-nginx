# Performance testing ingress-nginx-controller in GithubAction-CI
This README will evolve as the development of testing occurs.

## INFORMATION
### 1. No CPU/Memory for stress
- Github-Actions job runner is a 2core 7Gig VM so that limits what/how we test https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners

  <img width="863" alt="image" src="https://user-images.githubusercontent.com/5085914/169713673-7daa56ed-dffe-4c49-8600-4e9a01754d37.png">
  
  - Need to eventually get our own beefy runner, with enough cpu/memory to handle stress level load

### 2. Scale is work-in-progress
- We are grateful to have got a free account on K6.io, as part of their OSS Program. But it is limited to 600 tests per year.

### 3. No Testplans
- Testplan discussion and coding is needed for more practical real-world testing reports

## DESCRIPTION
 
### What
- An issue was created for performance tests, for the ingress-nginx-controller builds, https://github.com/kubernetes/ingress-nginx/issues/8033 .

### How
- A step by step guide to using https://k6.io with GithubActions is here https://k6.io/blog/load-testing-using-github-actions/

  - The link above contains sample code
    <img width="901" alt="image" src="https://user-images.githubusercontent.com/5085914/169714798-6ff62542-a591-4379-8f50-7678b6024936.png">


  - Copy sample test code from website and edit to taste
   <img width="646" alt="image" src="https://user-images.githubusercontent.com/5085914/169714937-1f6f2a86-36c0-4826-8e28-5e450d461353.png">

  - The CI launches a ubuntu environment and uses `make dev-env` to create a kind cluster. The popular https://httpbin.org api docker image is used to create a workload
 
    <img width="646" alt="image" src="https://user-images.githubusercontent.com/5085914/169714872-613ffd6a-36b5-4317-81fe-c39765a76649.png">

  - We don't want the test to block CI so this syntax from Github-Actions creates a button to run the test
    <img width="257" alt="image" src="https://user-images.githubusercontent.com/5085914/169715159-e7fbeada-fcb4-4c25-a65f-8be8547b7c19.png">

  - The button looks like this (the `Run Workflow` dropdown at bottom right of screenshot)
    <img width="1385" alt="image" src="https://user-images.githubusercontent.com/5085914/169715301-0752c5ed-9f84-4560-9a5e-c872c32041de.png">

    <img width="1380" alt="image" src="https://user-images.githubusercontent.com/5085914/169715321-8e76ee6b-2a85-4ed2-ba4e-9d8410518c42.png">


### fqdn
- Obtained a freenom domain `ingress-nginx-controller.ga` 

  - The test uses a fqdn `test.ingress-nginx-controller.ga`

  - The K6 api has configuration options for dns resolution of (above mentioned fqdn) to localhost/loopback/127.0.0.1 (`make dev-env` cluster)
    <img width="445" alt="image" src="https://user-images.githubusercontent.com/5085914/169716036-213d43c1-4801-4b4f-aee6-1c11c7812e6d.png">

  - Will need to discuss and decide on fqdn, as it relates to tls secret

### tls
- Procured a letsencrypt wildcard certificate for `*.ingress-nginx-controller.ga`

  - base64 encoded hash of the cert + key is stored in the `Github Project Settings Secrets` as a variable

  - The `GithubActions secrets` variables are decoded in the CI to create the TLS secret

    <img width="1250" alt="image" src="https://user-images.githubusercontent.com/5085914/169716088-030b9f6f-cdb1-470b-b10c-ea4a0fb8199f.png">


### Visualization
- Plan is to run tests locally on a kind cluster, in the CI pipeline, but push results to K6-cloud

  - Pushing and visualization on K6 cloud is as simple as executing `k6 run -o cloud test.js`

  - Currently there is a personal account in trial period (50 tests or 1 year limit) bing used

  - Pushing test-results from K6 tests on laptop, to K6-cloud personal trial account on K6-Cloud, to see what the graphs look like

    <img width="954" alt="image" src="https://user-images.githubusercontent.com/5085914/169713896-2cc3b775-38d9-43c6-8792-2a43329a8cfb.png">

    <img width="950" alt="image" src="https://user-images.githubusercontent.com/5085914/169713941-1671426d-9356-4c50-956b-b4003df4aa78.png">

  - The cli result looks like this
    <img width="835" alt="image" src="https://user-images.githubusercontent.com/5085914/169715209-68aa116a-020b-4f2d-8c8e-ec2c5f68b7b0.png">

- Before merging the PR, the testing is being done on personal Github project with exact same code as this PR here https://github.com/longwuyuan/k6-loadtest-example/runs/6545706269?check_suite_focus=true

