# Installation Using Helm Chart

!!! info
    Only Helm v3 is supported

## Installing NGINX Ingress Controller

NGINX Ingress controller can be installed via [Helm](https://helm.sh/) using the chart from the project repository.

#### Configuring Chart Repo
```shell
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
```

#### Installing Via Configured Repo
```shell
helm install ingress-nginx ingress-nginx/ingress-nginx
```

!!! tip
    You can use any release name 
    ```shell
    helm install <release-name> ingress-nginx/ingress-nginx
    ```

## Customised Install

NGINX Ingress Controller helm chart supports a log of config values using which one can configure the underlying NGINX ingress controller.

### Configuration Values

#### Controller Specific Values

| Field        | Description      | Default    | Type    |
|:------------:|:----------------:|:----------:|:-------:|

#### Common Values

| Field        | Description      | Default    | Type    |
|:------------:|:----------------:|:----------:|:-------:|
