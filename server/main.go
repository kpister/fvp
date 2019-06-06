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
	"sync"
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
	NodesState        map[string]*fvp.SendMsg_State // map node id to node state
	NodesQuorumSlices map[string][][]string         // map node id to quorum slices
	StateCounter      int32                         // track the number of transition of this node
	NodesFvpClients   map[string]fvp.ServerClient
	NodesAddrs        []string          // list of all neighbors in quorum slices
	Dictionary        map[string]string // the kv store
	IsEvil            bool
	Strategy          string
	QsSlices          [][]string // set for cfg file
	Term              int32

	BroadcastTimeout int
}

func (o *node) broadcast() {
	gl.Lock()
	n := deepcopy(o)
	gl.Unlock()

	if !n.IsEvil {

		// build arguments, list of states
		ks := make([]*fvp.SendMsg_State, 0)
		for _, state := range n.NodesState {
			temp := state // for some wierd go thing
			ks = append(ks, temp)
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

func (o *node) updateStates(states []*fvp.SendMsg_State) {
	for _, state := range states {

		if prevState, ok := n.NodesState[state.Id]; !ok {
			n.NodesState[state.Id] = state
			Log(n.Term, "send", "updating state of "+state.Id+" for first time")
		} else if state.Counter > prevState.Counter {
			n.NodesState[state.Id] = state
			Log(n.Term, "send", "updating state of "+state.Id+" for updated counter")
		}
	}
}

func (o *node) updateQuorumSlices(states []*fvp.SendMsg_State) {
	for _, state := range states {
		if _, ok := n.NodesQuorumSlices[state.Id]; !ok {
			n.NodesQuorumSlices[state.Id] = convertQuorumSlices(state.QuorumSlices)
		}
	}
}

// get a map of a statement to a list of nodes that voted for/accepted it
func (o *node) getStatements() (map[string][]string, map[string][]string) {
	gl.Lock()
	n := deepcopy(o)
	gl.Unlock()

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

func (o *node) Send(ctx context.Context, in *fvp.SendMsg) (*fvp.EmptyMessage, error) {
	gl.Lock()
	n := deepcopy(o)
	gl.Unlock()

	if in.Term != n.Term {
		return &fvp.EmptyMessage{}, nil
	}

	n.updateStates(in.KnownStates)
	n.updateQuorumSlices(in.KnownStates)

	votedForStmt2Nodes, acceptedStmt2Nodes := n.getStatements()

	for stmt, nodes := range votedForStmt2Nodes {
		if canVote(stmt, n.NodesState[n.ID].VotedFor) && canVote(stmt, n.NodesState[n.ID].Accepted) && !inArray(n.NodesState[n.ID].VotedFor, stmt) {
			Log(n.Term, "put", stmt+" voted")
			n.NodesState[n.ID].VotedFor = append(n.NodesState[n.ID].VotedFor, stmt)
			nodes = append(nodes, n.ID)
		}

		if inArray(nodes, n.ID) && n.checkQuorum(nodes) && !inArray(n.NodesState[n.ID].Accepted, stmt) {
			Log(n.Term, "put", stmt+" accept")
			n.NodesState[n.ID].Accepted = append(n.NodesState[n.ID].Accepted, stmt)
		}
	}

	for stmt, nodes := range acceptedStmt2Nodes {
		if canVote(stmt, n.NodesState[n.ID].Accepted) && canVote(stmt, n.NodesState[n.ID].VotedFor) && !inArray(n.NodesState[n.ID].VotedFor, stmt) {
			Log(n.Term, "put", stmt+" voted")
			n.NodesState[n.ID].VotedFor = append(n.NodesState[n.ID].VotedFor, stmt)
		}

		if !canVote(stmt, n.NodesState[n.ID].Accepted) {
			// maybe stuck?
			Log(n.Term, "send", "can't accept "+stmt)
			continue
		}

		if inArray(nodes, n.ID) && n.checkQuorum(nodes) {
			if !inArray(n.NodesState[n.ID].Confirmed, stmt) {
				Log(n.Term, "put", stmt+" end")
				n.NodesState[n.ID].Confirmed = append(n.NodesState[n.ID].Confirmed, stmt)
				stmt_pieces := strings.Split(stmt, "=")
				n.Dictionary[stmt_pieces[0]] = stmt_pieces[1]
			}
		} else if n.checkBlocking(nodes) { // assert statement is not in confict with any others we have voted for?
			if !inArray(n.NodesState[n.ID].Accepted, stmt) {
				Log(n.Term, "put", stmt+" accept")
				n.NodesState[n.ID].Accepted = append(n.NodesState[n.ID].Accepted, stmt)
			}
		}
	}

	return &fvp.EmptyMessage{}, nil
}

var (
	n          node
	configFile = flag.String("config", "cfg.json", "the file to read the configuration from")
	print      = flag.Bool("print", false, "to print the state")
	help       = flag.Bool("h", false, "for usage")
	gl         sync.Mutex
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

	go grpcServer.Serve(lis)

	ticker := time.NewTicker(time.Duration(n.BroadcastTimeout) * time.Millisecond)
	for range ticker.C {
		// spew.Dump(n.NodesState)
		if *print {
			prettyPrintMap(n.NodesState)
		}
		n.broadcast()
	}
}
