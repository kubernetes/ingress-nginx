package cmd

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Access to docker client",
}

var dockerListCmd = &cobra.Command{
	Use:   "list",
	Short: "list out all docker images",
	Long:  "List out all docker images available to us",
	Run: func(cmd *cobra.Command, args []string) {
		dockerList()

	},
}

var dockerBuildCmd = &cobra.Command{
	Use: "build",
	Short: "build a docker container"
	Long: "build a docker container for use with ingress-nginx"
	Run: func(cmd *cobra.Command, args []string){
		dockerBuild()
	},

}
func init() {
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerListCmd)
}

func dockerList() {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	}
}

func dockerBuild() error {
/*
     docker build \
         ${PLATFORM_FLAG} ${PLATFORM} \
--no-cache \
         --pull \
         --build-arg BASE_IMAGE="$(BASE_IMAGE)" \
         --build-arg VERSION="$(TAG)" \
         --build-arg TARGETARCH="$(ARCH)" \
         --build-arg COMMIT_SHA="$(COMMIT_SHA)" \
         --build-arg BUILD_ID="$(BUILD_ID)" \
         -t $(REGISTRY)/controller:$(TAG) rootfs
*/

     cli, err := client.NewClientWithOpts(client.FromEnv)
     if err != nil {
         panic(err)
     }

	 builder := io.Reader{}

	 options := docker.ImageCreateOptions{


	 }
	 buildReponse, err := cli.ImageBuild(context.Background(), builder, options)
	 if err != nil{
		 return err
	 }
	 return nil

 }
