package auth

import "testing"

func TestLoadSpec(t *testing.T) {
	spec := NewSpec("serviceName", "http://127.0.0.1:3000/_spec")
	spec.Load()
}
