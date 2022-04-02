package cmd

import (
	"os"

	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
)

var (
	rootCmd = &cobra.Command{
		Use:   "kube-webhook-certgen",
		Short: "Create certificates and patch them to admission hooks",
		Long: `Use this to create a ca and signed certificates and patch admission webhooks to allow for quick
	           installation and configuration of validating and admission webhooks.`,
		PreRun: configureLogging,
		Run:    rootCommand,
	}

	cfg = struct {
		logLevel           string
		logfmt             string
		secretName         string
		namespace          string
		certName           string
		keyName            string
		host               string
		apiServiceName     string
		webhookName        string
		patchValidating    bool
		patchMutating      bool
		patchFailurePolicy string
		kubeconfig         string
	}{}
)

// Execute is the main entry point for the program
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	filenameHook := filename.NewHook()
	filenameHook.Field = "source"
	log.AddHook(filenameHook)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
	rootCmd.Flags()
	rootCmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "info", "Log level: panic|fatal|error|warn|info|debug|trace")
	rootCmd.PersistentFlags().StringVar(&cfg.logfmt, "log-format", "json", "Log format: text|json")
	rootCmd.PersistentFlags().StringVar(&cfg.kubeconfig, "kubeconfig", "", "Path to kubeconfig file: e.g. ~/.kube/kind-config-kind")
}

func configureLogging(_ *cobra.Command, _ []string) {
	l, err := log.ParseLevel(cfg.logLevel)
	if err != nil {
		log.WithField("err", err).Fatal("Invalid error level")
	}
	log.SetLevel(l)
	log.SetFormatter(getFormatter(cfg.logfmt))
}

func rootCommand(cmd *cobra.Command, _ []string) {
	cmd.Help()
	os.Exit(1)
}

func getFormatter(logfmt string) log.Formatter {
	switch logfmt {
	case "json":
		return &log.JSONFormatter{}
	case "text":
		return &log.TextFormatter{}
	}

	log.Fatalf("invalid log format '%s'", logfmt)
	return nil
}

func newKubernetesClients(kubeconfig string) (kubernetes.Interface, clientset.Interface) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.WithError(err).Fatal("error building kubernetes config")
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.WithError(err).Fatal("error creating kubernetes client")
	}

	aggregatorClientset, err := clientset.NewForConfig(config)
	if err != nil {
		log.WithError(err).Fatal("error creating kubernetes aggregator client")
	}

	return c, aggregatorClientset
}
