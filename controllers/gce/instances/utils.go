package instances

import (
	compute "google.golang.org/api/compute/v1"

	"k8s.io/ingress/controllers/gce/utils"
)

// Helper method to create instance groups.
// This method exists to ensure that we are using the same logic at all places.
func EnsureInstanceGroupsAndPorts(nodePool NodePool, namer *utils.Namer, port int64) ([]*compute.InstanceGroup, *compute.NamedPort, error) {
	return nodePool.AddInstanceGroup(namer.IGName(), port)
}
