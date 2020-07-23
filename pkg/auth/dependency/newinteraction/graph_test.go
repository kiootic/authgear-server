package newinteraction_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func TestGraph(t *testing.T) {
	Convey("Graph.Accept", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		any := gomock.Any()

		type input1 struct{}
		type input2 struct{}
		type input3 struct{}
		type input4 struct{}

		// A --> B --> D
		//  |    ^
		//  |    |
		//  \--> C --\
		//       ^---/
		nodeA := NewMockNode(ctrl)
		nodeB := NewMockNode(ctrl)
		nodeC := NewMockNode(ctrl)
		nodeD := NewMockNode(ctrl)
		edgeB := NewMockEdge(ctrl)
		edgeC := NewMockEdge(ctrl)
		edgeD := NewMockEdge(ctrl)
		edgeE := NewMockEdge(ctrl)

		ctx := &newinteraction.Context{}
		g := &newinteraction.Graph{Nodes: []newinteraction.Node{nodeA}}

		nodeA.EXPECT().DeriveEdges(ctx, any).AnyTimes().Return(
			[]newinteraction.Edge{edgeB, edgeC}, nil,
		)
		nodeB.EXPECT().DeriveEdges(ctx, any).AnyTimes().Return(
			[]newinteraction.Edge{edgeD}, nil,
		)
		nodeC.EXPECT().DeriveEdges(ctx, any).AnyTimes().Return(
			[]newinteraction.Edge{edgeB, edgeE}, nil,
		)
		nodeD.EXPECT().DeriveEdges(ctx, any).AnyTimes().Return(
			[]newinteraction.Edge{}, nil,
		)

		edgeB.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *newinteraction.Context, g *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
				if _, ok := input.(input1); ok {
					return nodeB, nil
				}
				if _, ok := input.(input2); ok {
					return nodeB, nil
				}
				return nil, newinteraction.ErrIncompatibleInput
			})
		edgeC.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *newinteraction.Context, g *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
				if _, ok := input.(input3); ok {
					return nodeC, nil
				}
				return nil, newinteraction.ErrIncompatibleInput
			})
		edgeD.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *newinteraction.Context, g *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
				if _, ok := input.(input2); ok {
					return nodeD, nil
				}
				return nil, newinteraction.ErrIncompatibleInput
			})
		edgeE.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *newinteraction.Context, g *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
				if _, ok := input.(input4); ok {
					return nil, newinteraction.ErrSameNode
				}
				return nil, newinteraction.ErrIncompatibleInput
			})

		Convey("should go to deepest node", func() {
			nodeB.EXPECT().Apply(ctx, any)
			graph, edges, err := g.Accept(ctx, input1{})
			So(err, ShouldBeError, newinteraction.ErrInputRequired)
			So(graph.Nodes, ShouldResemble, []newinteraction.Node{nodeA, nodeB})
			So(edges, ShouldResemble, []newinteraction.Edge{edgeD})

			nodeB.EXPECT().Apply(ctx, any)
			nodeD.EXPECT().Apply(ctx, any)
			graph, edges, err = g.Accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []newinteraction.Node{nodeA, nodeB, nodeD})
			So(edges, ShouldResemble, []newinteraction.Edge{})

			nodeC.EXPECT().Apply(ctx, any)
			graph, edges, err = g.Accept(ctx, input3{})
			So(err, ShouldBeError, newinteraction.ErrInputRequired)
			So(graph.Nodes, ShouldResemble, []newinteraction.Node{nodeA, nodeC})
			So(edges, ShouldResemble, []newinteraction.Edge{edgeB, edgeE})

			nodeB.EXPECT().Apply(ctx, any)
			nodeD.EXPECT().Apply(ctx, any)
			graph, edges, err = graph.Accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []newinteraction.Node{nodeA, nodeC, nodeB, nodeD})
			So(edges, ShouldResemble, []newinteraction.Edge{})
		})

		Convey("should process looping edge", func() {
			nodeC.EXPECT().Apply(ctx, any)
			graph, edges, err := g.Accept(ctx, input3{})
			So(err, ShouldBeError, newinteraction.ErrInputRequired)
			So(graph.Nodes, ShouldResemble, []newinteraction.Node{nodeA, nodeC})
			So(edges, ShouldResemble, []newinteraction.Edge{edgeB, edgeE})

			graph, edges, err = graph.Accept(ctx, input4{})
			So(err, ShouldBeError, newinteraction.ErrInputRequired)
			So(graph.Nodes, ShouldResemble, []newinteraction.Node{nodeA, nodeC})
			So(edges, ShouldResemble, []newinteraction.Edge{edgeB, edgeE})

			nodeB.EXPECT().Apply(ctx, any)
			nodeD.EXPECT().Apply(ctx, any)
			graph, edges, err = graph.Accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []newinteraction.Node{nodeA, nodeC, nodeB, nodeD})
			So(edges, ShouldResemble, []newinteraction.Edge{})
		})
	})
}