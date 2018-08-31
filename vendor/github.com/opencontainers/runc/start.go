// +build linux

package main

import (
	"os"
	"runtime"

	"github.com/codegangsta/cli"
	"github.com/coreos/go-systemd/activation"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// default action is to start a container
var startCommand = cli.Command{
	Name:  "start",
	Usage: "create and run a container",
	ArgsUsage: `<container-id>

Where "<container-id>" is your name for the instance of the container that you
are starting. The name you provide for the container instance must be unique on
your host.`,
	Description: `The start command creates an instance of a container for a bundle. The bundle
is a directory with a specification file named "` + specConfig + `" and a root
filesystem.

The specification file includes an args parameter. The args parameter is used
to specify command(s) that get run when the container is started. To change the
command(s) that get executed on start, edit the args parameter of the spec. See
"runc spec --help" for more explanation.`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "bundle, b",
			Value: "",
			Usage: `path to the root of the bundle directory, defaults to the current directory`,
		},
		cli.StringFlag{
			Name:  "console",
			Value: "",
			Usage: "specify the pty slave path for use with the container",
		},
		cli.BoolFlag{
			Name:  "detach,d",
			Usage: "detach from the container's process",
		},
		cli.StringFlag{
			Name:  "pid-file",
			Value: "",
			Usage: "specify the file to write the process id to",
		},
		cli.BoolFlag{
			Name:  "no-subreaper",
			Usage: "disable the use of the subreaper used to reap reparented processes",
		},
		cli.BoolFlag{
			Name:  "no-pivot",
			Usage: "do not use pivot root to jail process inside rootfs.  This should be used whenever the rootfs is on top of a ramdisk",
		},
	},
	Action: func(context *cli.Context) {
		bundle := context.String("bundle")
		if bundle != "" {
			if err := os.Chdir(bundle); err != nil {
				fatal(err)
			}
		}
		spec, err := loadSpec(specConfig)
		if err != nil {
			fatal(err)
		}

		notifySocket := os.Getenv("NOTIFY_SOCKET")
		if notifySocket != "" {
			setupSdNotify(spec, notifySocket)
		}

		if os.Geteuid() != 0 {
			fatalf("runc should be run as root")
		}

		status, err := startContainer(context, spec)
		if err != nil {
			fatal(err)
		}
		// exit with the container's exit status so any external supervisor is
		// notified of the exit with the correct exit status.
		os.Exit(status)
	},
}

func init() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()
	}
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: `initialize the namespaces and launch the process (do not call it outside of runc)`,
	Action: func(context *cli.Context) {
		factory, _ := libcontainer.New("")
		if err := factory.StartInitialization(); err != nil {
			// as the error is sent back to the parent there is no need to log
			// or write it to stderr because the parent process will handle this
			os.Exit(1)
		}
		panic("libcontainer: container init failed to exec")
	},
}

func startContainer(context *cli.Context, spec *specs.Spec) (int, error) {
	id := context.Args().First()
	if id == "" {
		return -1, errEmptyID
	}
	container, err := createContainer(context, id, spec)
	if err != nil {
		return -1, err
	}
	detach := context.Bool("detach")
	// Support on-demand socket activation by passing file descriptors into the container init process.
	listenFDs := []*os.File{}
	if os.Getenv("LISTEN_FDS") != "" {
		listenFDs = activation.Files(false)
	}
	r := &runner{
		enableSubreaper: !context.Bool("no-subreaper"),
		shouldDestroy:   true,
		container:       container,
		listenFDs:       listenFDs,
		console:         context.String("console"),
		detach:          detach,
		pidFile:         context.String("pid-file"),
	}
	return r.run(&spec.Process)
}
