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

package ingresses

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"k8s.io/ingress-nginx/cmd/plugin/request"
	"k8s.io/ingress-nginx/cmd/plugin/util"
)

// CreateCommand creates and returns this cobra subcommand
func CreateCommand(flags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
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
	cmd.Flags().String("host", "", "Show just the ingress definitions for this hostname")
	cmd.Flags().Bool("all-namespaces", false, "Find ingress definitions from all namespaces")

	return cmd
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
		fmt.Fprintln(printer, "NAMESPACE\tINGRESS NAME\tHOST+PATH\tADDRESSES\tTLS\tSERVICE\tSERVICE PORT\tENDPOINTS")
	} else {
		fmt.Fprintln(printer, "INGRESS NAME\tHOST+PATH\tADDRESSES\tTLS\tSERVICE\tSERVICE PORT\tENDPOINTS")
	}

	for _, row := range rows {
		var tlsMsg string
		if row.TLS {
			tlsMsg = "YES"
		} else {
			tlsMsg = "NO"
		}

		numEndpoints, err := request.GetNumEndpoints(flags, row.Namespace, row.ServiceName)
		if err != nil {
			return err
		}
		if numEndpoints == nil {
			row.NumEndpoints = "N/A"
		} else {
			row.NumEndpoints = fmt.Sprint(*numEndpoints)
		}

		if allNamespaces {
			fmt.Fprintf(printer, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n", row.Namespace, row.IngressName, row.Host+row.Path, row.Address, tlsMsg, row.ServiceName, row.ServicePort, row.NumEndpoints)
		} else {
			fmt.Fprintf(printer, "%v\t%v\t%v\t%v\t%v\t%v\t%v\n", row.IngressName, row.Host+row.Path, row.Address, tlsMsg, row.ServiceName, row.ServicePort, row.NumEndpoints)
		}
	}

	return nil
}

type ingressRow struct {
	Namespace    string
	IngressName  string
	Host         string
	Path         string
	TLS          bool
	ServiceName  string
	ServicePort  string
	Address      string
	NumEndpoints string
}

func getIngressRows(ingresses *[]networking.Ingress) []ingressRow {
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
		if ing.Spec.DefaultBackend != nil {
			name, port := serviceToNameAndPort(ing.Spec.DefaultBackend.Service)
			defaultBackendService = name
			defaultBackendPort = port.String()
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
				svcName, svcPort := serviceToNameAndPort(path.Backend.Service)
				row := ingressRow{
					Namespace:   ing.Namespace,
					IngressName: ing.Name,
					Host:        rule.Host,
					Path:        path.Path,
					TLS:         hasTLS,
					ServiceName: svcName,
					ServicePort: svcPort.String(),
					Address:     address,
				}

				rows = append(rows, row)
			}
		}
	}

	return rows
}

func serviceToNameAndPort(svc *networking.IngressServiceBackend) (string, intstr.IntOrString) {
	var svcName string
	if svc != nil {
		svcName = svc.Name
		if svc.Port.Number > 0 {
			return svcName, intstr.FromInt(int(svc.Port.Number))
		}
		if svc.Port.Name != "" {
			return svcName, intstr.FromString(svc.Port.Name)
		}
	}
	return "", intstr.IntOrString{}
}
