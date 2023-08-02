package k8sclient

import (
	"sync"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	client "k8s.io/client-go/kubernetes"
)

var (
	once         sync.Once
	globalClient *client.Clientset
)

func GlobalClient(flags *genericclioptions.ConfigFlags) *client.Clientset {
	once.Do(func() {
		rawConfig, err := flags.ToRESTConfig()
		if err != nil {
			panic(err)
		}
		globalClient, err = client.NewForConfig(rawConfig)
		if err != nil {
			panic(err)
		}
	})
	return globalClient
}
