package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
)

func init() {
	newinteraction.RegisterIntent(&IntentRemoveIdentity{})
}

type IntentRemoveIdentity struct {
	UserID string `json:"user_id"`
}

func NewIntentRemoveIdentity(userID string) *IntentRemoveIdentity {
	return &IntentRemoveIdentity{
		UserID: userID,
	}
}

func (i *IntentRemoveIdentity) InstantiateRootNode(ctx *newinteraction.Context, graph *newinteraction.Graph) (newinteraction.Node, error) {
	edge := nodes.EdgeUseUser{UseUserID: i.UserID}
	return edge.Instantiate(ctx, graph, i)
}

func (i *IntentRemoveIdentity) DeriveEdgesForNode(ctx *newinteraction.Context, graph *newinteraction.Graph, node newinteraction.Node) ([]newinteraction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeUseUser:
		return []newinteraction.Edge{
			&nodes.EdgeRemoveIdentity{},
		}, nil
	case *nodes.NodeRemoveIdentity:
		return []newinteraction.Edge{
			&nodes.EdgeRemoveAuthenticator{
				IdentityInfo: node.IdentityInfo,
			},
		}, nil
	case *nodes.NodeRemoveAuthenticator:
		return nil, nil
	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}