package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/codeskyblue/go-sh"
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

	Long: "build a docker container for use with ingress-nginx",
	Run: func(cmd *cobra.Command, args []string) {
		dockerBuild()
	},
}

var dco dockerBuildOpts

func init() {
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerListCmd)
	dockerCmd.AddCommand(dockerBuildCmd)
	dockerBuildCmd.Flags().StringVar(&dco.PlatformFlag, "platformflag", "", "Setting the Docker --platform build flag")
	dockerBuildCmd.Flags().StringSliceVar(&dco.Platform, "platforms", PLATFORMS, "comma seperated list of platforms to build for container image")
	dockerBuildCmd.Flags().StringVar(&dco.BuildArgs.BaseImage, "base", "", "base image to build container off of")
	dockerBuildCmd.Flags().StringVar(&dco.Path, "path", "", "container build path")
	dockerBuildCmd.Flags().StringVar(&dco.BuildArgs.Version, "version", "", "docker tag to build")
	dockerBuildCmd.Flags().StringVar(&dco.BuildArgs.TargetArch, "targetarch", "", "target arch to build")
	dockerBuildCmd.Flags().StringVar(&dco.BuildArgs.CommitSHA, "commitsha", "", "build arg commit sha to add to build")
	dockerBuildCmd.Flags().StringVar(&dco.BuildArgs.BuildId, "build-id", "", "build id to add to container metadata")
	dockerBuildCmd.Flags().StringVar(&dco.DockerFile, "dockerfile", "", "dockerfile of image to build")
	dockerBuildCmd.Flags().StringVar(&dco.Image.Name, "name", "", "container image name registry/name:tag@digest")
	dockerBuildCmd.Flags().StringVar(&dco.Image.Registry, "registry", Registry, "Registry to tag image and push container registry/name:tag@digest")
	dockerBuildCmd.Flags().StringVar(&dco.Image.Tag, "tag", "", "container tag registry/name:tag@digest")
	dockerBuildCmd.Flags().StringVar(&dco.Image.Digest, "digest", "", "digest of container image registry/name:tag@digest")
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

var PLATFORMS = []string{"amd64", "arm", "arm64", "s390x"}
var BUILDX_PLATFORMS = "linux/amd64,linux/arm,linux/arm64,linux/s390x"
var Registry = "gcr.io/k8s-staging-ingress-nginx"
var PKG = "k8s.io/ingress-nginx"

type dockerBuildOpts struct {
	PlatformFlag string
	Platform     []string
	BuildArgs    BuildArgs
	DockerFile   string
	Image        Image
	Path         string
}

type BuildArgs struct {
	BaseImage  string
	Version    string
	TargetArch string
	CommitSHA  string
	BuildId    string
}

// ControllerImage - struct with info about controllers
type Image struct {
	Tag      string
	Digest   string
	Registry string
	Name     string
}

func (i Image) print() string {
	return fmt.Sprintf("%s/%s:%s@sha256:%s", i.Registry, i.Name, i.Tag, i.Digest)
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
	session := sh.NewSession()
	session.ShowCMD = true

	fmt.Printf("Container Build Path: %v\n", dco.Path)

	buildArgs(&dco.BuildArgs)

	fmt.Printf("Base image: %s\n", dco.BuildArgs.BaseImage)

	fmt.Printf("Build Args: %s\n", buildArgs)

	session.Command("docker", "build", "--no-cache", "--pull",
		"--build-arg", "BASE_IMAGE="+dco.BuildArgs.BaseImage,
		"--build-arg", "VERSION="+dco.BuildArgs.Version,
		"--build-arg", "TARGETARCH="+dco.BuildArgs.TargetArch,
		"--build-arg", "COMMIT_SHA="+dco.BuildArgs.CommitSHA,
		"--build-arg", "BUILD_ID="+dco.BuildArgs.BuildId,
		dco.Path).Run()

	return nil

}

func buildArgs(b *BuildArgs) {

	if b.BaseImage == "" {
		base, err := getIngressNginxBase()
		CheckIfError(err, "Issue Retrieving base image")
		b.BaseImage = base
	}

	if b.Version == "" {
		b.Version = "1.0.0-dev"
	}
	if b.TargetArch == "" {
		b.TargetArch = getArch()
	}
	if b.CommitSHA == "" {
		sha, _ := sh.Command("git", "rev-parse", "--short", "HEAD").Output()
		b.CommitSHA = strings.TrimSpace(string(sha))
	}
	if b.BuildId == "" {
		b.BuildId = "UNSET"
	}
}

func getIngressNginxBase() (string, error) {
	dat, err := os.ReadFile("../NGINX_BASE")
	CheckIfError(err, "Could not read NGINX_BASE file")
	datString := string(dat)
	//remove newline
	datString = strings.Replace(datString, "\n", "", -1)
	return datString, nil
}
