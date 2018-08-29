package main

import (
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/opencontainers/runc/libcontainer"
)

var deleteCommand = cli.Command{
	Name:  "delete",
	Usage: "delete any resources held by the container often used with detached containers",
	ArgsUsage: `<container-id>

Where "<container-id>" is the name for the instance of the container.
	 
For example, if the container id is "ubuntu01" and runc list currently shows the
status of "ubuntu01" as "destroyed" the following will delete resources held for
"ubuntu01" removing "ubuntu01" from the runc list of containers:  
	 
       # runc delete ubuntu01`,
	Action: func(context *cli.Context) {
		container, err := getContainer(context)
		if err != nil {
			if lerr, ok := err.(libcontainer.Error); ok && lerr.Code() == libcontainer.ContainerNotExists {
				// if there was an aborted start or something of the sort then the container's directory could exist but
				// libcontainer does not see it because the state.json file inside that directory was never created.
				path := filepath.Join(context.GlobalString("root"), context.Args().First())
				if err := os.RemoveAll(path); err == nil {
					return
				}
			}
			fatal(err)
		}
		destroy(container)
	},
}
