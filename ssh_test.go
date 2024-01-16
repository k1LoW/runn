package runn

import (
	"testing"

	"github.com/k1LoW/runn/testutil"
)

func TestNewSSHRunner(t *testing.T) {
	addr := testutil.SSHServer(t)
	_, err := newSSHRunner("sc", addr)
	if err != nil {
		t.Fatal(err)
	}
}
