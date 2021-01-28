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
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/ingress-nginx/internal/nginx"
)

const (
	backendsPath = "/configuration/backends"
	generalPath  = "/configuration/general"
	certsPath    = "/configuration/certs"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dbg",
		Short: "dbg is a tool for quickly inspecting the state of the nginx instance",
	}

	backendsCmd := &cobra.Command{
		Use:   "backends",
		Short: "Inspect the dynamically-loaded backends information",
	}
	rootCmd.AddCommand(backendsCmd)

	backendsAllCmd := &cobra.Command{
		Use:   "all",
		Short: "Output the all dynamic backend information as a JSON array",
		Run: func(cmd *cobra.Command, args []string) {
			backendsAll()
		},
	}
	backendsCmd.AddCommand(backendsAllCmd)

	backendsListCmd := &cobra.Command{
		Use:   "list",
		Short: "Output a newline-separated list of the backend names",
		Run: func(cmd *cobra.Command, args []string) {
			backendsList()
		},
	}
	backendsCmd.AddCommand(backendsListCmd)

	backendsGetCmd := &cobra.Command{
		Use:   "get [backend name]",
		Short: "Output the backend information only for the backend that has this name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			backendsGet(args[0])
		},
	}
	backendsCmd.AddCommand(backendsGetCmd)

	certCmd := &cobra.Command{
		Use:   "certs",
		Short: "Inspect dynamic SSL certificates",
	}

	certGetCmd := &cobra.Command{
		Use:   "get [hostname]",
		Short: "Get the dynamically-loaded certificate information for the given hostname",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			certGet(args[0])
			return nil
		},
	}
	certCmd.AddCommand(certGetCmd)

	rootCmd.AddCommand(certCmd)

	generalCmd := &cobra.Command{
		Use:   "general",
		Short: "Output the general dynamic lua state",
		Run: func(cmd *cobra.Command, args []string) {
			general()
		},
	}
	rootCmd.AddCommand(generalCmd)

	confCmd := &cobra.Command{
		Use:   "conf",
		Short: "Dump the contents of /etc/nginx/nginx.conf",
		Run: func(cmd *cobra.Command, args []string) {
			readNginxConf()
		},
	}
	rootCmd.AddCommand(confCmd)

	rootCmd.PersistentFlags().IntVar(&nginx.StatusPort, "status-port", 10246, `Port to use for the lua HTTP endpoint configuration.`)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func backendsAll() {
	statusCode, body, requestErr := nginx.NewGetStatusRequest(backendsPath)
	if requestErr != nil {
		fmt.Println(requestErr)
		return
	}
	if statusCode != 200 {
		fmt.Printf("Nginx returned code %v\n", statusCode)
		return
	}

	var prettyBuffer bytes.Buffer
	indentErr := json.Indent(&prettyBuffer, body, "", "  ")
	if indentErr != nil {
		fmt.Println(indentErr)
		return
	}

	fmt.Println(prettyBuffer.String())
}

func backendsList() {
	statusCode, body, requestErr := nginx.NewGetStatusRequest(backendsPath)
	if requestErr != nil {
		fmt.Println(requestErr)
		return
	}
	if statusCode != 200 {
		fmt.Printf("Nginx returned code %v\n", statusCode)
		return
	}

	var f interface{}
	unmarshalErr := json.Unmarshal(body, &f)
	if unmarshalErr != nil {
		fmt.Println(unmarshalErr)
		return
	}
	backends := f.([]interface{})

	for _, backendi := range backends {
		backend := backendi.(map[string]interface{})
		fmt.Println(backend["name"].(string))
	}
}

func backendsGet(name string) {
	statusCode, body, requestErr := nginx.NewGetStatusRequest(backendsPath)
	if requestErr != nil {
		fmt.Println(requestErr)
		return
	}
	if statusCode != 200 {
		fmt.Printf("Nginx returned code %v\n", statusCode)
		return
	}

	var f interface{}
	unmarshalErr := json.Unmarshal(body, &f)
	if unmarshalErr != nil {
		fmt.Println(unmarshalErr)
		return
	}
	backends := f.([]interface{})

	for _, backendi := range backends {
		backend := backendi.(map[string]interface{})
		if backend["name"].(string) == name {
			printed, _ := json.MarshalIndent(backend, "", "  ")
			fmt.Println(string(printed))
			return
		}
	}
	fmt.Println("A backend of this name was not found.")
}

func certGet(host string) {
	statusCode, body, requestErr := nginx.NewGetStatusRequest(certsPath + "?hostname=" + host)
	if requestErr != nil {
		fmt.Println(requestErr)
		return
	}

	if statusCode == 200 {
		fmt.Print(string(body))
		return
	} else if statusCode != 404 {
		fmt.Printf("Nginx returned code %v\n", statusCode)
		fmt.Println(string(body))
		return
	}

	fmt.Printf("No cert found for host %v\n", host)
}

func general() {
	//TODO: refactor to obtain ingress-nginx pod count from the api server
	/*
		statusCode, body, requestErr := nginx.NewGetStatusRequest(generalPath)
		if requestErr != nil {
			fmt.Println(requestErr)
			return
		}
		if statusCode != 200 {
			fmt.Printf("Nginx returned code %v\n", statusCode)
			return
		}
	*/

	var prettyBuffer bytes.Buffer
	indentErr := json.Indent(&prettyBuffer, []byte("{}"), "", "  ")
	if indentErr != nil {
		fmt.Println(indentErr)
		return
	}

	fmt.Println(prettyBuffer.String())
}

func readNginxConf() {
	conf, err := nginx.ReadNginxConf()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(conf)
}
