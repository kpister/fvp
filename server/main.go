package main

import (
	fvp "github.com/kpister/fvp/server/proto/fvp"
	kv "github.com/kpister/fvp/server/proto/kvstore"

	"context"
	"time"
)

type node struct {
	id                string
	nodesState        map[string]fvp.SendMsg_State
	nodesQuorumSlices map[string][][]string
	stateCounter      int32
}

func (n *node) broadcast() {

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

// getAllVotedStatements returns all the statements anyone is voting for
func (n *node) getAllVotedStatements() []string {
	votedStatements := make([]string, 0)
	for _, state := range n.nodesState {
		for _, vote := range state.VotedFor {
			if !inArray(votedStatements, vote) {
				votedStatements = append(votedStatements, vote)
			}
		}

	}
	return votedStatements
}

// getAllAcceptedStatements returns all the statements anyone is accepting
func (n *node) getAllAcceptedStatements() []string {
	acceptedStatements := make([]string, 0)
	for _, state := range n.nodesState {
		for _, accept := range state.VotedFor {
			if !inArray(acceptedStatements, accept) {
				acceptedStatements = append(acceptedStatements, accept)
			}
		}

	}
	return acceptedStatements
}

// check if the given list of nodes forms a quorum
// Def. for all nodes v in the given set of nodes U,
// there exists a quorum slice q of node v such that q is a subset of U
func (n *node) checkQuorum(nodes []string) bool {
	for _, node := range nodes {
		// find if there is a quorum slice of node v is a subset of U
		existed := false
		for _, quorumSlice := range n.nodesQuorumSlices[node] {
			isSubset := true
			for _, qsNode := range quorumSlice {
				found := false
				for _, node := range nodes {
					if qsNode == node {
						found = true
						break
					}
				}
				if !found {
					// this quorum slice is not a subset of U
					isSubset = false
					break
				}
			}
			if isSubset {
				// this quorum slice is a subset of U
				existed = true
				break
			}
		}
		if !existed {
			// no quorum slice of node v is a subset of U
			return false
		}
	}

	return true
}

// check if the given list of nodes forms a blocking set
// Def. for all quorum slices q of node v
// the intersection of q and the given set of nodes B is not an empty set
func (n *node) checkBlocking(nodes []string) bool {
	// for every quorum slice of the local node
	for _, quorumSlice := range n.nodesQuorumSlices[n.id] {
		// check if there exists a node in the list belonging to the quorum slice
		existed := false
		for _, node := range nodes {
			for _, qsNode := range quorumSlice {
				if node == qsNode {
					existed = true
					break
				}
			}
			if existed {
				break
			}
		}
		if !existed {
			// the intersectino of a quorum slice of v and the given set of nodes is an empty set
			return false
		}
	}

	return true
}

// checkVoteQuorum check if we have a quorum for any voted statement
func (n *node) checkVoteQuorum() []string {
	// assume that the state is updated
	votedStatements := n.getAllVotedStatements()
	// for each statement in votedStatements we need to check if a quorum of nodes are voting for it
	statementsWithQuorum := make([]string, 0)
	for _, statement := range votedStatements {
		if len(n.checkQuorumForVoteStatement(statement)) != 0 {
			// we have a quorum for this statement
			statementsWithQuorum = append(statementsWithQuorum, statement)
		}
	}

	// in correct protocol len(statementsWithQuorum) should atmost be 1
	return statementsWithQuorum

}

// checkQuorumForVoteStatement checks if we have a quorum which is voting for a particular statement
func (n *node) checkQuorumForVoteStatement(statement string) []string {
	nodesQSlices := make(map[string][][]string)

	// copy the quorum slices
	for k, v := range n.nodesQuorumSlices {
		nodesQSlices[k] = v
	}

	// for all nodes remove all the slices in which any node doesn't vote for the statement
	for k := range nodesQSlices {
		QSlices := nodesQSlices[k]
		l := len(QSlices)

		for i := 0; i < l; i++ {
			slice := QSlices[i]
			if !isUnanimousVote(slice, statement, n.nodesState) {
				QSlices = remove(QSlices, i)
				l--
			}
		}
		nodesQSlices[k] = QSlices
	}

	for {
		// look for nodes which now have empty quorum slice
		emptyNodes := make([]string, 0)
		for k := range nodesQSlices {
			QSlices := nodesQSlices[k]
			if len(QSlices) == 0 {
				emptyNodes = append(emptyNodes, k)
			}
		}

		if len(emptyNodes) == 0 {
			break
		}

		// remove all the slices with empty node
		for _, emptyNode := range emptyNodes {
			for k := range nodesQSlices {
				QSlices := nodesQSlices[k]
				l := len(QSlices)

				for i := 0; i < l; i++ {
					slice := QSlices[i]
					if inArray(slice, emptyNode) {
						QSlices = remove(QSlices, i)
						l--
					}
				}
				nodesQSlices[k] = QSlices
			}
		}
	}

	// check if we have any node with non-empty quorum slice set
	nodesInQuorum := make([]string, 0)
	for k := range nodesQSlices {
		QSlices := nodesQSlices[k]
		if len(QSlices) != 0 {
			nodesInQuorum = append(nodesInQuorum, k)
		}
	}
	return nodesInQuorum
}

func (n *node) Send(ctx context.Context, in *fvp.SendMsg) (*fvp.EmptyMessage, error) {
	votedForStmt2Nodes, acceptedStmt2Nodes := n.getStatements()

	for _, nodes := range votedForStmt2Nodes {
		n.checkBlocking(nodes)
		// n.checkQuorum(nodes)
		n.checkVoteQuorum()
	}

	for _, nodes := range acceptedStmt2Nodes {
		n.checkBlocking(nodes)
		// n.checkQuorum(nodes)
		n.checkVoteQuorum()
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

func createNode(nodeId string) *node {
	return &node{
		id:                nodeId,
		nodesState:        make(map[string]fvp.SendMsg_State, 0),
		nodesQuorumSlices: make(map[string][][]string, 0),
	}
}

func main() {
	setupLog("~/node_id/log.txt")
	// create node
	n := createNode("0")

	// setup grpc
	// ...

	// create timer
	ticker := time.NewTicker(50 * time.Millisecond)

	for range ticker.C {
		go n.broadcast()
	}
}
