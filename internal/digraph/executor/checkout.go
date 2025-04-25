package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dagu-org/dagu/internal/digraph"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-viper/mapstructure/v2"
)

const (
	gitCheckOutExecutorType = "git-checkout"
	accessTokenUser         = "dagu"
)

var (
	errGitCheckoutAuthConfigInvalid = fmt.Errorf("invalid git checkout auth config")
)

var _ Executor = (*gitCheckout)(nil)

func init() {
	Register(gitCheckOutExecutorType, newCheckout)
}

type gitCheckoutExecConfigDefinition struct {
	Repo       string                              `json:"repo"`
	Ref        string                              `json:"ref"`
	Path       string                              `json:"path"`
	Depth      int                                 `json:"depth"`
	Submodules bool                                `json:"submodules"`
	LFS        bool                                `json:"lfs"`
	Clean      bool                                `json:"clean"`
	Cache      bool                                `json:"cache"`
	Auth       gitCheckoutExecAuthConfigDefinition `json:"auth"`
}

func (g *gitCheckoutExecConfigDefinition) getRepoCachePath() string {
	// https://github.com/dagu-org/dagu.git -> github.com/dagu-org/dagu.git
	// git@github.com:dagu-org/dagu.git -> github.com/dagu-org/dagu.git

	repo := strings.TrimPrefix(g.Repo, "https://")
	repo = strings.TrimPrefix(repo, "git@")

	return fmt.Sprintf("~/.cache/dagu/git/%s", repo)
}

type gitCheckoutExecAuthConfigDefinition struct {
	TokenEnv       string `json:"tokenEnv"`
	UserName       string `json:"userName"`
	Password       string `json:"password"`
	SSHKey         string `json:"sshKey"`
	SSHKeyPassword string `json:"sshKeyPassword"`
	SSHAgent       bool   `json:"sshAgent"`
}

// httpAuthMethod returns http auth method, if auth config is not set, return nil
func (g *gitCheckoutExecAuthConfigDefinition) httpAuthMethod() (transport.AuthMethod, error) {
	if g.TokenEnv != "" {
		g.UserName = accessTokenUser
		g.Password = os.Getenv(g.TokenEnv)
	}

	if len(g.UserName) == 0 || len(g.Password) == 0 {
		return nil, errGitCheckoutAuthConfigInvalid
	}

	return &gitHttp.BasicAuth{
		Username: g.UserName,
		Password: g.Password,
	}, nil
}

// sshAuthMethod returns ssh auth method, if auth config is not set, return nil
func (g *gitCheckoutExecAuthConfigDefinition) sshAuthMethod() (transport.AuthMethod, error) {
	var (
		authMethod transport.AuthMethod
		publicKey  *ssh.PublicKeys
		err        error
	)

	if g.SSHAgent {
		if authMethod, err = ssh.NewSSHAgentAuth(accessTokenUser); err != nil {
			return nil, fmt.Errorf("failed to create ssh agent auth: %w", err)
		}

		return authMethod, nil
	}

	if _, err = os.Stat(g.SSHKey); err != nil {
		return nil, fmt.Errorf("failed to find ssh key file: %w", err)
	}

	if publicKey, err = ssh.NewPublicKeysFromFile(accessTokenUser, g.SSHKey, g.SSHKeyPassword); err != nil {
		return nil, fmt.Errorf("failed to create ssh public keys: %w", err)
	}

	return publicKey, nil
}

func (g *gitCheckoutExecAuthConfigDefinition) authMethod(repo string) (transport.AuthMethod, error) {
	// example: https://github.com/dagu-org/dagu.git, use http auth
	if strings.HasPrefix(repo, "https://") {
		return g.httpAuthMethod()
	}

	// example: git@github.com:dagu-org/dagu.git, use ssh auth
	return g.sshAuthMethod()
}

type gitCheckoutExecConfig struct {
	repo          string
	ref           string
	path          string
	depth         int
	submodules    bool
	lfs           bool
	clean         bool
	cache         bool
	repoCachePath string
}

type gitCheckout struct {
	stdout     io.Writer
	stderr     io.Writer
	authMethod transport.AuthMethod
	config     *gitCheckoutExecConfig
}

func convertFromDef(def *gitCheckoutExecConfigDefinition, authMethod transport.AuthMethod) *gitCheckout {
	return &gitCheckout{
		stdout: os.Stdout,
		stderr: os.Stderr,
		config: &gitCheckoutExecConfig{
			repo:          def.Repo,
			ref:           def.Ref,
			path:          def.Path,
			depth:         def.Depth,
			submodules:    def.Submodules,
			lfs:           def.LFS,
			clean:         def.Clean,
			cache:         def.Cache,
			repoCachePath: def.getRepoCachePath(),
		},
		authMethod: authMethod,
	}
}

func newCheckout(_ context.Context, step digraph.Step) (Executor, error) {
	var (
		def        = &gitCheckoutExecConfigDefinition{}
		authMethod transport.AuthMethod
		err        error
	)

	if err = decodeGitCheckoutConfig(step.ExecutorConfig.Config, def); err != nil {
		return nil, fmt.Errorf("failed to decode git checkout config: %w", err)
	}

	if authMethod, err = def.Auth.authMethod(def.Repo); err != nil {
		return nil, err
	}

	return convertFromDef(def, authMethod), nil
}

func decodeGitCheckoutConfig(data map[string]any, config *gitCheckoutExecConfigDefinition) error {
	var (
		mapDecoder   *mapstructure.Decoder
		decodeConfig = &mapstructure.DecoderConfig{
			Result:           config,
			WeaklyTypedInput: true,
		}
		err error
	)

	if mapDecoder, err = mapstructure.NewDecoder(decodeConfig); err != nil {
		return fmt.Errorf("failed to create map decoder: %w", err)
	}

	if err = mapDecoder.Decode(data); err != nil {
		return fmt.Errorf("failed to decode git checkout config: %w", err)
	}

	return nil
}

func (g *gitCheckout) SetStdout(out io.Writer) {
	g.stdout = out
}

func (g *gitCheckout) SetStderr(out io.Writer) {
	g.stderr = out
}

func (g *gitCheckout) Kill(sig os.Signal) error {
	//TODO implement me
	panic("implement me")
}

// initCacheMirror initializes the cache mirror for the git repository while enabling cache.
func (g *gitCheckout) initCacheMirror() error {
	var (
		err error
	)

	if !g.config.cache {
		return nil
	}

	if _, err = os.Stat(g.config.repoCachePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check cache mirror: %w", err)
		}

		if err = os.MkdirAll(g.config.repoCachePath, 0755); err != nil {
			return fmt.Errorf("failed to create cache mirror: %w", err)
		}
	}

	return nil
}

func (g *gitCheckout) getGitCloneOptions() *git.CloneOptions {
	var (
		options = &git.CloneOptions{
			URL:      g.config.repo,
			Auth:     g.authMethod,
			Progress: g.stdout,
			Depth:    g.config.depth,
		}
	)

	if g.config.submodules {
		options.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	}

	if g.config.lfs {
		options.SingleBranch = true
		options.ReferenceName = plumbing.NewRemoteReferenceName("origin", "HEAD")
	}

	if g.config.cache {
		options.ReferenceName = plumbing.NewRemoteReferenceName(g.config.repoCachePath, "HEAD")
	}

	return options
}

func (g *gitCheckout) Run(ctx context.Context) error {
	var (
		err error
	)

	if err = g.initCacheMirror(); err != nil {
		return err
	}

	if _, err = git.PlainCloneContext(ctx, g.config.path, false, g.getGitCloneOptions()); err != nil {
		return fmt.Errorf("failed to clone git repository: %w", err)
	}

	return nil
}
