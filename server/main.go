package main

import (
	fvp "github.com/kpister/fvp/server/proto/fvp"
	kv "github.com/kpister/fvp/server/proto/kvstore"

	"context"
	"time"
)

type node struct {
	Id                string
	nodesState        map[string]fvp.SendMsg_State
	nodesQuorumSlices map[string][][]string
	stateCounter      int32
}

func (n *node) broadcast() {
	/*
	   for n' in neighbors {
	       n'.send(n.state)
	   }
	*/
}

func (n *node) updateStates(states []*fvp.SendMsg_State) {
	for _, state := range states {
		if prevState, ok := n.nodesState[state.Id]; !ok {
			n.nodesState[state.Id] = *state
		} else if state.Counter > prevState.Counter {
			n.nodesState[state.Id] = *state
		}
	}
}

func (n *node) updateQuorumSlices(states []*fvp.SendMsg_State) {
	for _, state := range states {
		n.nodesQuorumSlices[state.Id] = convertQuorumSlices(state.QuorumSlices)
	}
}

func convertQuorumSlices(qs []*fvp.SendMsg_Slice) [][]string {
	return make([][]string, 0)
}

// get a map of a statement to a list of nodes that voted for/accepted it
func (n *node) getStatements() (map[string][]string, map[string][]string) {
	votedForStmt2Nodes := make(map[string][]string, 0)
	acceptedStmt2Nodes := make(map[string][]string, 0)
	for node, state := range n.nodesState {
		for _, statement := range state.VotedFor {
			votedForStmt2Nodes[statement] = append(votedForStmt2Nodes[statement], node)
		}
		for _, statement := range state.Accepted {
			acceptedStmt2Nodes[statement] = append(acceptedStmt2Nodes[statement], node)
		}
	}

	return votedForStmt2Nodes, acceptedStmt2Nodes
}

// check if the given list of nodes forms a blocking set
func (n *node) checkBlocking(nodes []string) bool {
	// for every quorum slice of the local node
	isBlockingSet := true
	for _, quorumSlice := range n.nodesQuorumSlices[n.Id] {
		// check if there exists a node in the list belonging to the quorum slice
		existed := false
		for _, node := range nodes {
			for _, qsNode := range quorumSlice {
				if node == qsNode {
					existed = true
					break
				}
			}
		}
		if !existed {
			isBlockingSet = false
			break
		}
	}

	return isBlockingSet
}

// check if the given list of nodes forms a quorum
func (n *node) checkQuorum(nodes []string) bool {
	return true
}

func (n *node) Send(ctx context.Context, in *fvp.SendMsg) (*fvp.EmptyMessage, error) {
	votedForStmt2Nodes, acceptedStmt2Nodes := n.getStatements()

	for _, nodes := range votedForStmt2Nodes {
		n.checkBlocking(nodes)
		n.checkQuorum(nodes)
	}

	for _, nodes := range acceptedStmt2Nodes {
		n.checkBlocking(nodes)
		n.checkQuorum(nodes)
	}

	// transition(n.state)

	return &fvp.EmptyMessage{}, nil
}

func (n *node) Get(ctx context.Context, in *kv.GetRequest) (*kv.GetResponse, error) {

	// lookup in.key in dictionary

	return &kv.GetResponse{Value: "", Ret: kv.ReturnCode_SUCCESS}, nil
}

func (n *node) Put(ctx context.Context, in *kv.PutRequest) (*kv.PutResponse, error) {

	// set value in dictionary

	return &kv.PutResponse{Ret: kv.ReturnCode_SUCCESS}, nil
}

func createNode() *node {
	return &node{
		nodesState:        make(map[string]fvp.SendMsg_State, 0),
		nodesQuorumSlices: make(map[string][][]string, 0),
	}
}

func main() {
	setupLog("~/node_id/log.txt")
	// create node
	n := createNode()

	// setup grpc
	// ...

	// create timer
	ticker := time.NewTicker(50 * time.Millisecond)

	for range ticker.C {
		go n.broadcast()
	}
}
