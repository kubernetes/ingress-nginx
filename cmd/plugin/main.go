/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"text/tabwriter"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	//Just importing this is supposed to allow cloud authentication
	// eg GCP, AWS, Azure ...
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/ingress-nginx/cmd/plugin/request"
	"k8s.io/ingress-nginx/cmd/plugin/util"
	"k8s.io/ingress-nginx/internal/nginx"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ingress-nginx",
		Short: "A kubectl plugin for inspecting your ingress-nginx deployments",
	}

	// Respect some basic kubectl flags like --namespace
	flags := genericclioptions.NewConfigFlags()
	flags.AddFlags(rootCmd.PersistentFlags())

	ingCmd := &cobra.Command{
		Use:     "ingresses",
		Aliases: []string{"ingress", "ing"},
		Short:   "Provide a short summary of all of the ingress definitions",
		RunE: func(cmd *cobra.Command, args []string) error {
			host, err := cmd.Flags().GetString("host")
			if err != nil {
				return err
			}

			allNamespaces, err := cmd.Flags().GetBool("all-namespaces")
			if err != nil {
				return err
			}

			util.PrintError(ingresses(flags, host, allNamespaces))
			return nil
		},
	}
	ingCmd.Flags().String("host", "", "Show just the ingress definitions for this hostname")
	ingCmd.Flags().Bool("all-namespaces", false, "Find ingress definitions from all namespaces")
	rootCmd.AddCommand(ingCmd)

	confCmd := &cobra.Command{
		Use:   "conf",
		Short: "Inspect the generated nginx.conf",
		RunE: func(cmd *cobra.Command, args []string) error {
			host, err := cmd.Flags().GetString("host")
			if err != nil {
				return err
			}

			pod, err := cmd.Flags().GetString("pod")
			if err != nil {
				return err
			}

			util.PrintError(conf(flags, host, pod))
			return nil
		},
	}
	confCmd.Flags().String("host", "", "Print just the server block with this hostname")
	confCmd.Flags().String("pod", "", "Query a particular ingress-nginx pod")
	rootCmd.AddCommand(confCmd)

	generalCmd := &cobra.Command{
		Use:   "general",
		Short: "Inspect the other dynamic ingress-nginx information",
		RunE: func(cmd *cobra.Command, args []string) error {
			pod, err := cmd.Flags().GetString("pod")
			if err != nil {
				return err
			}

			util.PrintError(general(flags, pod))
			return nil
		},
	}
	generalCmd.Flags().String("pod", "", "Query a particular ingress-nginx pod")
	rootCmd.AddCommand(generalCmd)

	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Show information about the ingress-nginx service",
		RunE: func(cmd *cobra.Command, args []string) error {
			util.PrintError(info(flags))
			return nil
		},
	}
	rootCmd.AddCommand(infoCmd)

	backendsCmd := &cobra.Command{
		Use:   "backends",
		Short: "Inspect the dynamic backend information of an ingress-nginx instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			pod, err := cmd.Flags().GetString("pod")
			if err != nil {
				return err
			}
			backend, err := cmd.Flags().GetString("backend")
			if err != nil {
				return err
			}
			onlyList, err := cmd.Flags().GetBool("list")
			if err != nil {
				return err
			}
			if onlyList && backend != "" {
				return fmt.Errorf("--list and --backend cannot both be specified")
			}

			util.PrintError(backends(flags, pod, backend, onlyList))
			return nil
		},
	}
	backendsCmd.Flags().String("pod", "", "Query a particular ingress-nginx pod")
	backendsCmd.Flags().String("backend", "", "Output only the information for the given backend")
	backendsCmd.Flags().Bool("list", false, "Output a newline-separated list of backend names")
	rootCmd.AddCommand(backendsCmd)

	certsCmd := &cobra.Command{
		Use:   "certs",
		Short: "Output the certificate data stored in an ingress-nginx pod",
		RunE: func(cmd *cobra.Command, args []string) error {
			pod, err := cmd.Flags().GetString("pod")
			if err != nil {
				return err
			}
			host, err := cmd.Flags().GetString("host")
			if err != nil {
				return err
			}

			util.PrintError(certs(flags, pod, host))
			return nil
		},
	}
	certsCmd.Flags().String("host", "", "Get the cert for this hostname")
	certsCmd.Flags().String("pod", "", "Query a particular ingress-nginx pod")
	cobra.MarkFlagRequired(certsCmd.Flags(), "host")
	rootCmd.AddCommand(certsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func certs(flags *genericclioptions.ConfigFlags, pod string, host string) error {
	command := []string{"/dbg", "certs", "get", host}
	var out string
	var err error
	if pod != "" {
		out, err = request.NamedPodExec(flags, pod, command)
	} else {
		out, err = request.IngressPodExec(flags, command)
	}
	if err != nil {
		return err
	}

	fmt.Print(out)
	return nil
}

func info(flags *genericclioptions.ConfigFlags) error {
	service, err := request.GetIngressService(flags)
	if err != nil {
		return err
	}

	fmt.Printf("Service cluster IP address: %v\n", service.Spec.ClusterIP)
	fmt.Printf("LoadBalancer IP|CNAME: %v\n", service.Spec.LoadBalancerIP)
	return nil
}

func backends(flags *genericclioptions.ConfigFlags, pod string, backend string, onlyList bool) error {
	var command []string
	if onlyList {
		command = []string{"/dbg", "backends", "list"}
	} else if backend != "" {
		command = []string{"/dbg", "backends", "get", backend}
	} else {
		command = []string{"/dbg", "backends", "all"}
	}

	var out string
	var err error
	if pod != "" {
		out, err = request.NamedPodExec(flags, pod, command)
	} else {
		out, err = request.IngressPodExec(flags, command)
	}
	if err != nil {
		return err
	}

	fmt.Print(out)
	return nil
}

func general(flags *genericclioptions.ConfigFlags, pod string) error {
	var general string
	var err error
	if pod != "" {
		general, err = request.NamedPodExec(flags, pod, []string{"/dbg", "general"})
	} else {
		general, err = request.IngressPodExec(flags, []string{"/dbg", "general"})
	}
	if err != nil {
		return err
	}

	fmt.Print(general)
	return nil
}

func ingresses(flags *genericclioptions.ConfigFlags, host string, allNamespaces bool) error {
	var namespace string
	if allNamespaces {
		namespace = ""
	} else {
		namespace = util.GetNamespace(flags)
	}

	ingresses, err := request.GetIngressDefinitions(flags, namespace)
	if err != nil {
		return err
	}

	rows := getIngressRows(&ingresses)

	if host != "" {
		rowsWithHost := make([]ingressRow, 0)
		for _, row := range rows {
			if row.Host == host {
				rowsWithHost = append(rowsWithHost, row)
			}
		}
		rows = rowsWithHost
	}

	printer := tabwriter.NewWriter(os.Stdout, 6, 4, 3, ' ', 0)
	defer printer.Flush()

	if allNamespaces {
		fmt.Fprintln(printer, "NAMESPACE\tINGRESS NAME\tHOST+PATH\tADDRESSES\tTLS\tSERVICE\tSERVICE PORT")
	} else {
		fmt.Fprintln(printer, "INGRESS NAME\tHOST+PATH\tADDRESSES\tTLS\tSERVICE\tSERVICE PORT")
	}

	for _, row := range rows {
		var tlsMsg string
		if row.TLS {
			tlsMsg = "YES"
		} else {
			tlsMsg = "NO"
		}
		if allNamespaces {
			fmt.Fprintf(printer, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t\n", row.Namespace, row.IngressName, row.Host+row.Path, row.Address, tlsMsg, row.ServiceName, row.ServicePort)
		} else {
			fmt.Fprintf(printer, "%v\t%v\t%v\t%v\t%v\t%v\t\n", row.IngressName, row.Host+row.Path, row.Address, tlsMsg, row.ServiceName, row.ServicePort)
		}
	}

	return nil
}

func conf(flags *genericclioptions.ConfigFlags, host string, pod string) error {
	var nginxConf string
	var err error
	if pod != "" {
		nginxConf, err = request.NamedPodExec(flags, pod, []string{"/dbg", "conf"})
	} else {
		nginxConf, err = request.IngressPodExec(flags, []string{"/dbg", "conf"})
	}
	if err != nil {
		return err
	}

	if host != "" {
		block, err := nginx.GetServerBlock(nginxConf, host)
		if err != nil {
			return err
		}

		fmt.Println(strings.TrimRight(strings.Trim(block, " \n"), " \n\t"))
	} else {
		fmt.Print(nginxConf)
	}

	return nil
}

type ingressRow struct {
	Namespace   string
	IngressName string
	Host        string
	Path        string
	TLS         bool
	ServiceName string
	ServicePort string
	Address     string
}

func getIngressRows(ingresses *[]v1beta1.Ingress) []ingressRow {
	rows := make([]ingressRow, 0)

	for _, ing := range *ingresses {

		address := ""
		for _, lbIng := range ing.Status.LoadBalancer.Ingress {
			if len(lbIng.IP) > 0 {
				address = address + lbIng.IP + ","
			}
			if len(lbIng.Hostname) > 0 {
				address = address + lbIng.Hostname + ","
			}
		}
		if len(address) > 0 {
			address = address[:len(address)-1]
		}

		tlsHosts := make(map[string]struct{})
		for _, tls := range ing.Spec.TLS {
			for _, host := range tls.Hosts {
				tlsHosts[host] = struct{}{}
			}
		}

		defaultBackendService := ""
		defaultBackendPort := ""
		if ing.Spec.Backend != nil {
			defaultBackendService = ing.Spec.Backend.ServiceName
			defaultBackendPort = ing.Spec.Backend.ServicePort.String()
		}

		// Handle catch-all ingress
		if len(ing.Spec.Rules) == 0 && len(defaultBackendService) > 0 {
			row := ingressRow{
				Namespace:   ing.Namespace,
				IngressName: ing.Name,
				Host:        "*",
				ServiceName: defaultBackendService,
				ServicePort: defaultBackendPort,
				Address:     address,
			}

			rows = append(rows, row)
			continue
		}

		for _, rule := range ing.Spec.Rules {
			_, hasTLS := tlsHosts[rule.Host]

			//Handle ingress with no paths
			if rule.HTTP == nil {
				row := ingressRow{
					Namespace:   ing.Namespace,
					IngressName: ing.Name,
					Host:        rule.Host,
					Path:        "",
					TLS:         hasTLS,
					ServiceName: defaultBackendService,
					ServicePort: defaultBackendPort,
					Address:     address,
				}
				rows = append(rows, row)
				continue
			}

			for _, path := range rule.HTTP.Paths {
				row := ingressRow{
					Namespace:   ing.Namespace,
					IngressName: ing.Name,
					Host:        rule.Host,
					Path:        path.Path,
					TLS:         hasTLS,
					ServiceName: path.Backend.ServiceName,
					ServicePort: path.Backend.ServicePort.String(),
					Address:     address,
				}

				rows = append(rows, row)
			}
		}
	}

	return rows
}
