package main

import (
	"flag"

	fvp "github.com/kpister/fvp/server/proto/fvp"
	kv "github.com/kpister/fvp/server/proto/kvstore"

	"context"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"google.golang.org/grpc"
)

/*
* TODO: consider non-unanimous votes in quorum slice based on threshold % -- inf -- .
* TODO: write monitor getstate -- 3 -- everyone
* TODO: test, test, test -- 5 -- everyone
* TODO: run benchmarks -- last -- everyone
* TODO: write report
 */

type node struct {
	ID                string
	NodesState        map[string]fvp.SendMsg_State // map node id to node state
	NodesQuorumSlices map[string][][]string        // map node id to quorum slices
	StateCounter      int32                        // track the number of transition of this node
	NodesFvpClients   map[string]fvp.ServerClient
	NodesAddrs        []string          // list of all neighbors in quorum slices
	Dictionary        map[string]string // the kv store
	IsEvil            bool
	Strategy          string
	QsSlices          [][]string // set for cfg file
	Term              int32

	BroadcastTimeout int
}

func (n *node) broadcast() {
	if !n.IsEvil {

		// build arguments, list of states
		ks := make([]*fvp.SendMsg_State, 0)
		for _, state := range n.NodesState {
			temp := state
			ks = append(ks, &temp)
		}
		args := &fvp.SendMsg{KnownStates: ks, Term: n.Term}

		// for every neighbor send the message

		for _, neighbor := range n.NodesAddrs {
			ctx := context.Background()

			_, err := n.NodesFvpClients[neighbor].Send(ctx, args)
			if err != nil {
				n.errorHandler(err, "broadcast", neighbor)
			}
		}

	} else {
		n.evilBehavior(n.Strategy)
	}
}

func (n *node) updateStates(states []*fvp.SendMsg_State) {
	for _, state := range states {

		if prevState, ok := n.NodesState[state.Id]; !ok {
			n.NodesState[state.Id] = *state
			Log(n.Term, "send", "updating state of "+state.Id+" for first time")
		} else if state.Counter > prevState.Counter {
			n.NodesState[state.Id] = *state
			Log(n.Term, "send", "updating state of "+state.Id+" for updated counter")
		}
	}
}

func (n *node) updateQuorumSlices(states []*fvp.SendMsg_State) {
	for _, state := range states {
		if _, ok := n.NodesQuorumSlices[state.Id]; !ok {
			n.NodesQuorumSlices[state.Id] = convertQuorumSlices(state.QuorumSlices)
		}
	}
}

// get a map of a statement to a list of nodes that voted for/accepted it
func (n *node) getStatements() (map[string][]string, map[string][]string) {
	votedForStmt2Nodes := make(map[string][]string, 0)
	acceptedStmt2Nodes := make(map[string][]string, 0)
	for node, state := range n.NodesState {
		for _, statement := range state.VotedFor {
			votedForStmt2Nodes[statement] = append(votedForStmt2Nodes[statement], node)
		}
		for _, statement := range state.Accepted {
			if !inArray(votedForStmt2Nodes[statement], node) {
				votedForStmt2Nodes[statement] = append(votedForStmt2Nodes[statement], node)
			}
			acceptedStmt2Nodes[statement] = append(acceptedStmt2Nodes[statement], node)
		}
	}

	return votedForStmt2Nodes, acceptedStmt2Nodes
}

func (n *node) Send(ctx context.Context, in *fvp.SendMsg) (*fvp.EmptyMessage, error) {
	if in.Term != n.Term {
		return &fvp.EmptyMessage{}, nil
	}

	n.updateStates(in.KnownStates)
	n.updateQuorumSlices(in.KnownStates)

	votedForStmt2Nodes, acceptedStmt2Nodes := n.getStatements()

	update := false
	votedFor := n.NodesState[n.ID].VotedFor
	accepted := n.NodesState[n.ID].Accepted
	confirmed := n.NodesState[n.ID].Confirmed

	for stmt, nodes := range votedForStmt2Nodes {
		if canVote(stmt, votedFor) && canVote(stmt, accepted) && !inArray(votedFor, stmt) {
			Log(n.Term, "put", stmt+" voted")
			votedFor = append(votedFor, stmt)
			nodes = append(nodes, n.ID)
			update = true
		}
		if !inArray(nodes, n.ID) { // we are not in these nodes
			continue
		}
		if n.checkQuorum(nodes) {
			if !inArray(accepted, stmt) {
				Log(n.Term, "put", stmt+" accept")
				accepted = append(accepted, stmt)
				update = true
			}
		}
	}

	for stmt, nodes := range acceptedStmt2Nodes {
		if canVote(stmt, accepted) && canVote(stmt, votedFor) && !inArray(votedFor, stmt) {
			Log(n.Term, "put", stmt+" voted")
			votedFor = append(votedFor, stmt)
			update = true
		}

		if !canVote(stmt, accepted) {
			// maybe stuck?
			Log(n.Term, "send", "can't accept "+stmt)
			continue
		}

		if !inArray(nodes, n.ID) { // we are not in these nodes
			continue
		}

		if n.checkQuorum(nodes) {
			if !inArray(confirmed, stmt) {
				Log(n.Term, "put", stmt+" end")
				confirmed = append(confirmed, stmt)
				update = true
			}
		} else if n.checkBlocking(nodes) { // assert statement is not in confict with any others we have voted for?
			if !inArray(accepted, stmt) {
				Log(n.Term, "put", stmt+" accept")
				accepted = append(accepted, stmt)
				update = true
			}
		}
	}

	if update {
		n.StateCounter++
		n.NodesState[n.ID] = fvp.SendMsg_State{
			Id:           n.ID,
			Accepted:     accepted,
			VotedFor:     votedFor,
			Confirmed:    confirmed,
			QuorumSlices: n.NodesState[n.ID].QuorumSlices,
			Counter:      n.StateCounter,
		}
	}

	return &fvp.EmptyMessage{}, nil
}

var (
	n          node
	configFile = flag.String("config", "cfg.json", "the file to read the configuration from")
	print      = flag.Bool("print", false, "to print the state")
	help       = flag.Bool("h", false, "for usage")
)

func main() {
	if n.IsEvil {
		setupLog(path.Join(os.Getenv("HOME"), "logs", n.ID+".evil"))
	} else {
		setupLog(path.Join(os.Getenv("HOME"), "logs", n.ID+".txt"))
	}

	// setup grpc
	lis, err := net.Listen("tcp", ":"+strings.Split(n.ID, ":")[1])
	if err != nil {
		Log(n.Term, "connection", fmt.Sprintf("Failed to listen on the port. %v", err))
	}

	grpcServer := grpc.NewServer()
	fvp.RegisterServerServer(grpcServer, &n)
	kv.RegisterKeyValueStoreServer(grpcServer, &n)
	n.buildClients()

	Log(n.Term, "connection", "Listening on "+n.ID)

	ticker := time.NewTicker(time.Duration(n.BroadcastTimeout) * time.Millisecond)
	go func() {
		for range ticker.C {
			// spew.Dump(n.NodesState)
			if *print {
				prettyPrintMap(n.NodesState)
			}
			go n.broadcast()
		}
	}()

	grpcServer.Serve(lis)
}
