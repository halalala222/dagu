package executor

import (
	"context"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"os"
	"testing"
)

func TestGitCheckout_Run(t *testing.T) {
	var (
		g = &gitCheckout{
			stdout: os.Stdout,
			stderr: os.Stderr,
			config: &gitCheckoutExecConfig{
				repo:  "git@github.com:halalala222/git-test.git",
				ref:   "main",
				path:  "dagu",
				depth: 1,
			},
		}
		authMethod transport.AuthMethod
		err        error
	)

	if authMethod, err = ssh.NewSSHAgentAuth("git"); err != nil {
		t.Errorf("failed to create auth method: %v", err)
	}
	g.authMethod = authMethod

	if err = g.Run(context.Background()); err != nil {
		t.Errorf("failed to run git checkout: %v", err)
	}
}
