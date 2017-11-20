package controller

import (
	"sync/atomic"

	extensions "k8s.io/api/extensions/v1beta1"
)

func (n *NGINXController) isForceReload() bool {
	return atomic.LoadInt32(&n.forceReload) != 0
}

// SetForceReload sets if the ingress controller should be reloaded or not
func (n *NGINXController) SetForceReload(shouldReload bool) {
	if shouldReload {
		atomic.StoreInt32(&n.forceReload, 1)
		n.syncQueue.Enqueue(&extensions.Ingress{})
		return
	}

	atomic.StoreInt32(&n.forceReload, 0)
}
