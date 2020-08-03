package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeUseIdentityLoginID{})
}

type InputUseIdentityLoginID interface {
	GetLoginIDKey() string
	GetLoginID() string
}

type UseIdentityLoginIDMode string

const (
	UseIdentityLoginIDModeCreate = "create"
	UseIdentityLoginIDModeSelect = "select"
	UseIdentityLoginIDModeUpdate = "update"
)

type EdgeUseIdentityLoginID struct {
	Mode    UseIdentityLoginIDMode
	Configs []config.LoginIDKeyConfig
}

// GetIdentityCandidates implements IdentityCandidatesGetter.
func (e *EdgeUseIdentityLoginID) GetIdentityCandidates() []identity.Candidate {
	candidates := make([]identity.Candidate, len(e.Configs))
	for i, c := range e.Configs {
		conf := c
		candidates[i] = identity.NewLoginIDCandidate(&conf)
	}
	return candidates
}

func (e *EdgeUseIdentityLoginID) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputUseIdentityLoginID)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	loginIDKey := input.GetLoginIDKey()
	loginID := input.GetLoginID()
	claims := map[string]interface{}{
		identity.IdentityClaimLoginIDKey:   loginIDKey,
		identity.IdentityClaimLoginIDValue: loginID,
	}
	spec := &identity.Spec{
		Type:   authn.IdentityTypeLoginID,
		Claims: claims,
	}

	return &NodeUseIdentityLoginID{
		Mode:         e.Mode,
		IdentitySpec: spec,
	}, nil

}

type NodeUseIdentityLoginID struct {
	Mode         UseIdentityLoginIDMode `json:"mode"`
	IdentitySpec *identity.Spec         `json:"identity_spec"`
}

func (n *NodeUseIdentityLoginID) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeUseIdentityLoginID) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	switch n.Mode {
	case UseIdentityLoginIDModeCreate:
		return []newinteraction.Edge{&EdgeCreateIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
	case UseIdentityLoginIDModeSelect:
		return []newinteraction.Edge{&EdgeSelectIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
	case UseIdentityLoginIDModeUpdate:
		return []newinteraction.Edge{&EdgeUpdateIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
	default:
		panic(fmt.Errorf("interaction: unexpected use identity mode: %v", n.Mode))
	}
}