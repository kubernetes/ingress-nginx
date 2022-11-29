## helm

The Helm package manager for Kubernetes.

### Synopsis

The Kubernetes package manager

Common actions for Helm:

- helm search:    search for charts
- helm pull:      download a chart to your local directory to view
- helm install:   upload the chart to Kubernetes
- helm list:      list releases of charts

Environment variables:

| Name                               | Description                                                                       |
|------------------------------------|-----------------------------------------------------------------------------------|
| $HELM_CACHE_HOME                   | set an alternative location for storing cached files.                             |
| $HELM_CONFIG_HOME                  | set an alternative location for storing Helm configuration.                       |
| $HELM_DATA_HOME                    | set an alternative location for storing Helm data.                                |
| $HELM_DEBUG                        | indicate whether or not Helm is running in Debug mode                             |
| $HELM_DRIVER                       | set the backend storage driver. Values are: configmap, secret, memory, sql.       |
| $HELM_DRIVER_SQL_CONNECTION_STRING | set the connection string the SQL storage driver should use.                      |
| $HELM_MAX_HISTORY                  | set the maximum number of helm release history.                                   |
| $HELM_NAMESPACE                    | set the namespace used for the helm operations.                                   |
| $HELM_NO_PLUGINS                   | disable plugins. Set HELM_NO_PLUGINS=1 to disable plugins.                        |
| $HELM_PLUGINS                      | set the path to the plugins directory                                             |
| $HELM_REGISTRY_CONFIG              | set the path to the registry config file.                                         |
| $HELM_REPOSITORY_CACHE             | set the path to the repository cache directory                                    |
| $HELM_REPOSITORY_CONFIG            | set the path to the repositories file.                                            |
| $KUBECONFIG                        | set an alternative Kubernetes configuration file (default "~/.kube/config")       |
| $HELM_KUBEAPISERVER                | set the Kubernetes API Server Endpoint for authentication                         |
| $HELM_KUBECAFILE                   | set the Kubernetes certificate authority file.                                    |
| $HELM_KUBEASGROUPS                 | set the Groups to use for impersonation using a comma-separated list.             |
| $HELM_KUBEASUSER                   | set the Username to impersonate for the operation.                                |
| $HELM_KUBECONTEXT                  | set the name of the kubeconfig context.                                           |
| $HELM_KUBETOKEN                    | set the Bearer KubeToken used for authentication.                                 |

Helm stores cache, configuration, and data based on the following configuration order:

- If a HELM_*_HOME environment variable is set, it will be used
- Otherwise, on systems supporting the XDG base directory specification, the XDG variables will be used
- When no other location is set a default location will be used based on the operating system

By default, the default directories depend on the Operating System. The defaults are listed below:

| Operating System | Cache Path                | Configuration Path             | Data Path               |
|------------------|---------------------------|--------------------------------|-------------------------|
| Linux            | $HOME/.cache/helm         | $HOME/.config/helm             | $HOME/.local/share/helm |
| macOS            | $HOME/Library/Caches/helm | $HOME/Library/Preferences/helm | $HOME/Library/helm      |
| Windows          | %TEMP%\helm               | %APPDATA%\helm                 | %APPDATA%\helm          |


### Options

```
      --debug                       enable verbose output
  -h, --help                        help for helm
      --kube-apiserver string       the address and the port for the Kubernetes API server
      --kube-as-group stringArray   group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --kube-as-user string         username to impersonate for the operation
      --kube-ca-file string         the certificate authority file for the Kubernetes API server connection
      --kube-context string         name of the kubeconfig context to use
      --kube-token string           bearer token used for authentication
      --kubeconfig string           path to the kubeconfig file
  -n, --namespace string            namespace scope for this request
      --registry-config string      path to the registry config file (default "/home/sjouke/.config/helm/registry/config.json")
      --repository-cache string     path to the file containing cached repository indexes (default "/home/sjouke/.cache/helm/repository")
      --repository-config string    path to the file containing repository names and URLs (default "/home/sjouke/.config/helm/repositories.yaml")
```

### SEE ALSO

* [helm cm-push](helm_cm-push.md)	 - Please see https://github.com/chartmuseum/helm-push for usage
* [helm completion](helm_completion.md)	 - generate autocompletion scripts for the specified shell
* [helm create](helm_create.md)	 - create a new chart with the given name
* [helm dependency](helm_dependency.md)	 - manage a chart's dependencies
* [helm env](helm_env.md)	 - helm client environment information
* [helm get](helm_get.md)	 - download extended information of a named release
* [helm history](helm_history.md)	 - fetch release history
* [helm install](helm_install.md)	 - install a chart
* [helm lint](helm_lint.md)	 - examine a chart for possible issues
* [helm list](helm_list.md)	 - list releases
* [helm package](helm_package.md)	 - package a chart directory into a chart archive
* [helm plugin](helm_plugin.md)	 - install, list, or uninstall Helm plugins
* [helm pull](helm_pull.md)	 - download a chart from a repository and (optionally) unpack it in local directory
* [helm push](helm_push.md)	 - push a chart to remote
* [helm registry](helm_registry.md)	 - login to or logout from a registry
* [helm repo](helm_repo.md)	 - add, list, remove, update, and index chart repositories
* [helm rollback](helm_rollback.md)	 - roll back a release to a previous revision
* [helm search](helm_search.md)	 - search for a keyword in charts
* [helm show](helm_show.md)	 - show information of a chart
* [helm status](helm_status.md)	 - display the status of the named release
* [helm template](helm_template.md)	 - locally render templates
* [helm test](helm_test.md)	 - run tests for a release
* [helm uninstall](helm_uninstall.md)	 - uninstall a release
* [helm upgrade](helm_upgrade.md)	 - upgrade a release
* [helm verify](helm_verify.md)	 - verify that a chart at the given path has been signed and is valid
* [helm version](helm_version.md)	 - print the client version information

###### Auto generated by spf13/cobra on 29-Nov-2022
