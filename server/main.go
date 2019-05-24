package main

import (
	fvp "github.com/kpister/fvp/server/proto/fvp"
	kv "github.com/kpister/fvp/server/proto/kvstore"

	"context"
	"time"
)

type node struct {
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
		n.nodesQuorumSlices[state.Id] = convertQuorumSlice(*state.QuorumSlices)
	}
}

func convertQuorumSlices(qs []*fvp.SendMsg_Slice) [][]string {
	return make([][]string, 0)
}

func (n *node) getStatements() {
	statement2votedNodes := make(map[string][]string)
	statement2acceptedNodes := make(map[string][]string)
	for _, state := range n.nodesState {

	}
}

func (n *node) getStatements() {
	// return of a union of all voted for accepted statements from each state
}

// check if all blocking set members have voted for or accepted a statement
func (n *node) checkBlocking() {
	// for every statement
	// for every n.quorumslices
	// check if the statement exists in that quorum slice
	// true for all
	// blocking
}

// check if all quorum members have for or accepted a statement
func (n *node) checkQuorum() {
}

func (n *node) Send(ctx context.Context, in *fvp.SendMsg) (*fvp.EmptyMessage, error) {

	n.check_blocking(in.KnownStates)
	// check_quorum(in.votedfor)

	// check_blocking(in.accepted)
	// check_quorum(in.accepted)

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
		nodesState: make([]*fvp.SendMsg_State, 0),
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
