package store

import (
	"fmt"

	"k8s.io/client-go/tools/cache"

	"k8s.io/ingress-nginx/internal/ingress"
)

// SSLCertTracker holds a store of referenced Secrets in Ingress rules
type SSLCertTracker struct {
	cache.ThreadSafeStore
}

// NewSSLCertTracker creates a new SSLCertTracker store
func NewSSLCertTracker() *SSLCertTracker {
	return &SSLCertTracker{
		cache.NewThreadSafeStore(cache.Indexers{}, cache.Indices{}),
	}
}

// GetByNamespaceName searches for an ingress in the local ingress Store
func (s SSLCertTracker) GetByNamespaceName(key string) (*ingress.SSLCert, error) {
	cert, exists := s.Get(key)
	if !exists {
		return nil, fmt.Errorf("local SSL certificate %v was not found", key)
	}
	return cert.(*ingress.SSLCert), nil
}
