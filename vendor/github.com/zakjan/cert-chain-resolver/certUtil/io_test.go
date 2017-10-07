package certUtil

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsPEM(t *testing.T) {
	assert.True(t, IsPEM([]byte("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----")))
	assert.False(t, IsPEM([]byte("xxx")))
}
