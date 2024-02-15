package runn

import (
	"testing"
)

func TestNewSSHRunner(t *testing.T) {
	_, err := newSSHRunner("sc", "localhost:22")
	if err != nil {
		t.Fatal(err)
	}
}
