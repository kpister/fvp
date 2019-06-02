package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"math/rand"

	fvp "github.com/kpister/fvp/server/proto/fvp"
	kv "github.com/kpister/fvp/server/proto/kvstore"
	mon "github.com/kpister/fvp/server/proto/monitor"

	"context"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
)

/*
* TODO: add json loading of configs -- 1 -- kaiser
* TODO: add error handler -- 1 -- Yu-Wun
* TODO: consider non-unanimous votes in quorum slice based on threshold % -- inf -- .
* TODO: write get/set client -- 2 -- Kaiser
* TODO: write monitor getstate -- 3 -- everyone
* TODO: setup remote -- 4 -- .
* TODO: test, test, test -- 5 -- everyone
* TODO: run benchmarks -- last -- everyone
* TODO: write report
 */

type node struct {
	ID                string
	NodesState        map[string]fvp.SendMsg_State
	NodesQuorumSlices map[string][][]string
	StateCounter      int32
	// NodesFvpClients   map[string]fvp.ServerClient
	NodesAddrs    map[string]string
	Dictionary    map[string]string
	MonitorClient mon.MonitorClient
}

func (n *node) getNeighbors() []string {
	neighbors := make([]string, 0)
	for _, slice := range n.NodesQuorumSlices[n.ID] {
		for _, neighbor := range slice {
			// don't add yourself
			if neighbor == n.ID {
				continue
			}
			// don't add twice
			if inArray(neighbors, neighbor) {
				continue
			}

			neighbors = append(neighbors, neighbor)
		}
	}
	return neighbors
}

func (n *node) broadcast() {

	// build arguments, list of states
	ks := make([]*fvp.SendMsg_State, 0)
	for _, state := range n.NodesState {
		ks = append(ks, &state)
	}
	sendmsg := &fvp.SendMsg{KnownStates: ks}

	neighbors := n.getNeighbors()

	args := &mon.MonitorSendMsg{
		From: n.ID,
		To:   neighbors,
		Msg:  sendmsg,
	}

	ctx, _ := context.WithTimeout(
		context.Background(),
		time.Duration(1000)*time.Millisecond)

	_, err := n.MonitorClient.MonitorSend(ctx, args)
	if err != nil {
		n.errorHandler(err, "broadcast", *monitorAddr)
	}

}

func (n *node) updateStates(states []*fvp.SendMsg_State) {
	for _, state := range states {

		if prevState, ok := n.NodesState[state.Id]; !ok {
			n.NodesState[state.Id] = *state
			Log("low", "send", "updating state of "+state.Id+"for first time")
		} else if state.Counter > prevState.Counter {
			n.NodesState[state.Id] = *state
			Log("low", "send", "updating state of "+state.Id+"for updated counter")
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

func convertQuorumSlices(qs []*fvp.SendMsg_Slice) [][]string {
	ret := make([][]string, 0)

	for _, el := range qs {
		ret = append(ret, el.Nodes)
	}
	return ret
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

// getAllVotedStatements returns all the statements anyone is voting for
func (n *node) getAllVotedStatements() []string {
	votedStatements := make([]string, 0)
	for _, state := range n.NodesState {
		for _, vote := range state.VotedFor {
			if !inArray(votedStatements, vote) {
				votedStatements = append(votedStatements, vote)
			}
		}
		for _, vote := range state.Accepted {
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
	for _, state := range n.NodesState {
		for _, accept := range state.Accepted {
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
		for _, quorumSlice := range n.NodesQuorumSlices[node] {
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
	for _, quorumSlice := range n.NodesQuorumSlices[n.ID] {
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
	for k, v := range n.NodesQuorumSlices {
		nodesQSlices[k] = v
	}

	// for all nodes remove all the slices in which any node doesn't vote for the statement
	for k := range nodesQSlices {
		QSlices := nodesQSlices[k]
		l := len(QSlices)

		for i := 0; i < l; i++ {
			slice := QSlices[i]
			if !isUnanimousVote(slice, statement, n.NodesState) {
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
		// we need to remove the node itself from nodesQSlices?
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
	// Log("low", "send", "Send received")
	n.updateStates(in.KnownStates)
	n.updateQuorumSlices(in.KnownStates)

	votedForStmt2Nodes, acceptedStmt2Nodes := n.getStatements()

	update := false
	accepted := n.NodesState[n.ID].Accepted
	confirmed := n.NodesState[n.ID].Confirmed
	for stmt, nodes := range votedForStmt2Nodes {
		if n.checkQuorum(nodes) {
			if !inArray(accepted, stmt) {
				accepted = append(accepted, stmt)
				update = true
			}
		}
	}

	for stmt, nodes := range acceptedStmt2Nodes {
		if n.checkQuorum(nodes) {
			if !inArray(confirmed, stmt) {
				confirmed = append(confirmed, stmt)
				update = true
			}
		} else if n.checkBlocking(nodes) { // assert statement is not in confict with any others we have voted for?
			if !inArray(accepted, stmt) {
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
			VotedFor:     n.NodesState[n.ID].VotedFor,
			Confirmed:    confirmed,
			QuorumSlices: n.NodesState[n.ID].QuorumSlices,
			Counter:      n.StateCounter,
		}
	}

	return &fvp.EmptyMessage{}, nil
}

func (n *node) Get(ctx context.Context, in *kv.GetRequest) (*kv.GetResponse, error) {
	// lookup in.key in dictionary
	if val, ok := n.Dictionary[in.Key]; ok {
		return &kv.GetResponse{Value: val, Ret: kv.ReturnCode_SUCCESS}, nil
	}

	return &kv.GetResponse{Value: "", Ret: kv.ReturnCode_FAILURE}, nil
}

func (n *node) Put(ctx context.Context, in *kv.PutRequest) (*kv.PutResponse, error) {
	// set value in dictionary
	n.Dictionary[in.Key] = in.Value

	return &kv.PutResponse{Ret: kv.ReturnCode_SUCCESS}, nil
}

func (n *node) createNode() {

	// n.NodesAddrs = make(map[string]string, 0)
	n.NodesState = make(map[string]fvp.SendMsg_State, 0)
	// n.NodesQuorumSlices = make(map[string][][]string, 0)
	// n.NodesFvpClients = make(map[string]fvp.ServerClient, 0)

	// append our own state to Nodesstate
	ourSlices := make([]*fvp.SendMsg_Slice, 0)
	for _, slice := range n.NodesQuorumSlices[n.ID] {
		s := &fvp.SendMsg_Slice{
			Nodes: slice,
		}
		ourSlices = append(ourSlices, s)
	}

	ourState := fvp.SendMsg_State{
		Accepted:     make([]string, 0),
		Confirmed:    make([]string, 0),
		Counter:      0,
		Id:           n.ID,
		QuorumSlices: ourSlices,
		VotedFor:     []string{getVote(0.6, "a", "b")},
	}
	n.NodesState[n.ID] = ourState

	// fmt.Println(n.NodesQuorumSlices)
	// fmt.Println(n.NodesAddrs)
	// fmt.Println(n.ID)

	n.Dictionary = make(map[string]string, 0)
}

func (n *node) connectMonitor() {
	// grpc will retry in 20 ms at most 5 times when failed
	opts := []grpc_retry.CallOption{
		grpc_retry.WithMax(5),
		grpc_retry.WithPerRetryTimeout(1000 * time.Millisecond),
	}

	// connect to the monitor
	conn, err := grpc.Dial(*monitorAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)))
	if err != nil {
		Log("low", "connection", fmt.Sprintf("Failed to connect to %s. %v\n", *monitorAddr, err))
	}
	n.MonitorClient = mon.NewMonitorClient(conn)
}

var (
	n           *node
	configFile  = flag.String("config", "cfg.json", "the file to read the configuration from")
	printmap    = flag.Bool("print", false, "prints state")
	help        = flag.Bool("h", false, "for usage")
	monitorAddr = flag.String("mon", "localhost:9000", "the address of the monitor")
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(1)
	}

	readConfig(*configFile)
}

func readConfig(configFile string) {
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(configData, &n)
	n.createNode()

	if err != nil {
		fmt.Println(err)
	}
}

func main() {

	// TODO: remove this. directly assign log path, id, and address for now

	// os.MkdirAll(path.Join(os.Getenv("HOME"), n.ID), 0755)
	// setupLog(path.Join(os.Getenv("HOME"), n.ID, "log.txt"))
	os.MkdirAll(path.Join(os.Getenv("LOCAL"), n.ID), 0755)
	setupLog(path.Join(os.Getenv("LOCAL"), n.ID, "log.txt"))

	// n.NodesAddrs[n.ID] = "localhost:8000"

	// setup grpc
	lis, err := net.Listen("tcp", ":"+strings.Split(n.ID, ":")[1])
	if err != nil {
		Log("low", "connection", fmt.Sprintf("Failed to listen on the port. %v", err))
	}

	grpcServer := grpc.NewServer()
	fvp.RegisterServerServer(grpcServer, n)
	kv.RegisterKeyValueStoreServer(grpcServer, n)
	n.connectMonitor()

	Log("low", "connection", "Listening on "+n.ID)

	ticker := time.NewTicker(2000 * time.Millisecond)
	go func() {
		for range ticker.C {
			// spew.Dump(n.NodesState)
			if *printmap {
				prettyPrintMap(n.NodesState)
			}
			go n.broadcast()
		}
	}()

	grpcServer.Serve(lis)
}
