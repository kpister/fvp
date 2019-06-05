package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"math/rand"

	fvp "github.com/kpister/fvp/server/proto/fvp"
	kv "github.com/kpister/fvp/server/proto/kvstore"

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
}

func (n *node) broadcast() {
	if !n.IsEvil {
		sent := make([]string, 0)

		// build arguments, list of states
		ks := make([]*fvp.SendMsg_State, 0)
		for _, state := range n.NodesState {
			ks = append(ks, &state)
		}
		args := &fvp.SendMsg{KnownStates: ks}

		// for every neighbor send the message
		for _, slice := range n.NodesQuorumSlices[n.ID] {
			for _, neighbor := range slice {

				// don't send to yourself
				if neighbor == n.ID {
					continue
				}

				// don't send twice
				if inArray(sent, neighbor) {
					continue
				}
				sent = append(sent, neighbor)

				// Log("low", "broadcast", "broadcasting to "+neighbor)
				// TODO add cancel?
				ctx, _ := context.WithTimeout(
					context.Background(),
					timeout)

				_, err := n.NodesFvpClients[neighbor].Send(ctx, args)
				if err != nil {
					n.errorHandler(err, "broadcast", neighbor)
				}
			}
		}
	} else {
		n.evilBehavior(n.Strategy)
	}
}

func (n *node) evilBehavior(strategy string) {
	// all good ndoes are doing a=1, we do a=2

	for _, addr := range n.NodesAddrs {
		// don't send to yourself
		if addr == n.ID {
			continue
		}

		var fakestate fvp.SendMsg_State
		if strategy == "tc1" || strategy == "tc2" || strategy == "tc_sybil" {
			fakestate = fvp.SendMsg_State{
				Id:           n.ID,
				Accepted:     []string{"a=2"},
				VotedFor:     []string{"a=2"},
				Confirmed:    []string{"a=2"},
				QuorumSlices: n.NodesState[n.ID].QuorumSlices,
				Counter:      n.StateCounter,
			}
		} else if strategy == "random" {
			// TODO genarate ranodm quorum slice
			randSlices := make([]*fvp.SendMsg_Slice, 0)

			fakestate = fvp.SendMsg_State{
				Id:           n.ID,
				Accepted:     []string{getVote(0.5, "a=2", "a=3")},
				VotedFor:     []string{getVote(0.5, "a=2", "a=3")},
				Confirmed:    []string{getVote(0.5, "a=2", "a=3")},
				QuorumSlices: randSlices,
				Counter:      n.StateCounter,
			}
		} else if strategy == "tc3" {
			if addr == "localhost:8001" || addr == "localhost:8002" {
				// send a=1 to one half
				fakestate = fvp.SendMsg_State{
					Id:           n.ID,
					Accepted:     []string{"a=1"},
					VotedFor:     []string{"a=1"},
					Confirmed:    []string{"a=1"},
					QuorumSlices: n.NodesState[n.ID].QuorumSlices,
					Counter:      n.StateCounter,
				}
			} else { // localhost:8004 || localhost:8005
				// send a=2 to other half
				fakestate = fvp.SendMsg_State{
					Id:           n.ID,
					Accepted:     []string{"a=2"},
					VotedFor:     []string{"a=2"},
					Confirmed:    []string{"a=2"},
					QuorumSlices: n.NodesState[n.ID].QuorumSlices,
					Counter:      n.StateCounter,
				}
			}
		} else if strategy == "tc4" {
			if addr == "localhost:8001" || addr == "localhost:8002" {
				// send a=1 to one half
				fakestate = fvp.SendMsg_State{
					Id:           n.ID,
					Accepted:     []string{"a=1"},
					VotedFor:     []string{"a=1"},
					Confirmed:    []string{"a=1"},
					QuorumSlices: n.NodesState[n.ID].QuorumSlices,
					Counter:      n.StateCounter,
				}
			} else if addr == "localhost:8005" || addr == "localhost:8006" {
				// send a=2 to other half
				fakestate = fvp.SendMsg_State{
					Id:           n.ID,
					Accepted:     []string{"a=2"},
					VotedFor:     []string{"a=2"},
					Confirmed:    []string{"a=2"},
					QuorumSlices: n.NodesState[n.ID].QuorumSlices,
					Counter:      n.StateCounter,
				}
			} else { // localhost:4
				// vote := getVote(0.5, "a=1", "a=2")
				vote := "a=1"
				fakestate = fvp.SendMsg_State{
					Id:           n.ID,
					Accepted:     []string{vote},
					VotedFor:     []string{vote},
					Confirmed:    []string{vote},
					QuorumSlices: n.NodesState[n.ID].QuorumSlices,
					Counter:      n.StateCounter,
				}
			}
		} else {
			Log("medium", "broadcast", "Invalid evil strategy")
		}

		// prettyPrintMap(n.NodesState)
		// build arguments, list of states
		ks := make([]*fvp.SendMsg_State, 0)
		for _, state := range n.NodesState {
			if state.Id != n.ID { // don't lie about other's states
				temp := state // for some wierd go thing
				ks = append(ks, &temp)
			} else {
				temp := fakestate // for some wierd go thing
				ks = append(ks, &temp)
			}
		}

		args := &fvp.SendMsg{KnownStates: ks}
		// fmt.Println("to", addr)
		// prettyPrintMap(args)
		ctx, _ := context.WithTimeout(
			context.Background(),
			timeout)

		_, err := n.NodesFvpClients[addr].Send(ctx, args)
		if err != nil {
			n.errorHandler(err, "broadcast", addr)
		}

	}

}

func (n *node) updateStates(states []*fvp.SendMsg_State) {
	for _, state := range states {

		if prevState, ok := n.NodesState[state.Id]; !ok {
			n.NodesState[state.Id] = *state
			Log("send", "updating state of "+state.Id+"for first time")
		} else if state.Counter > prevState.Counter {
			n.NodesState[state.Id] = *state
			Log("send", "updating state of "+state.Id+"for updated counter")
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

func canVote(stmt string, list []string) bool {
	// assert stmt key is not in list, or if it is stmt value = list[key]
	pieces := strings.Split(stmt, "=")
	// assert len(pieces) == 2
	stmt_key := pieces[0]
	stmt_value := pieces[1]
	for _, s_ := range list {
		// key=value
		pieces = strings.Split(s_, "=")
		key := pieces[0]
		value := pieces[1]

		if key == stmt_key && value != stmt_value {
			return false
		}
	}
	return true
}

func (n *node) Send(ctx context.Context, in *fvp.SendMsg) (*fvp.EmptyMessage, error) {
	n.updateStates(in.KnownStates)
	n.updateQuorumSlices(in.KnownStates)

	votedForStmt2Nodes, acceptedStmt2Nodes := n.getStatements()

	update := false
	votedFor := n.NodesState[n.ID].VotedFor
	accepted := n.NodesState[n.ID].Accepted
	confirmed := n.NodesState[n.ID].Confirmed

	for stmt, nodes := range votedForStmt2Nodes {
		if canVote(stmt, votedFor) && canVote(stmt, accepted) && !inArray(votedFor, stmt) {
			votedFor = append(votedFor, stmt)
			nodes = append(nodes, n.ID)
			update = true
		}
		if !inArray(nodes, n.ID) { // we are not in these nodes
			continue
		}
		if n.checkQuorum(nodes) {
			if !inArray(accepted, stmt) {
				Log("put", stmt+" accept")
				accepted = append(accepted, stmt)
				update = true
			}
		}
	}

	for stmt, nodes := range acceptedStmt2Nodes {
		if canVote(stmt, accepted) && canVote(stmt, votedFor) && !inArray(votedFor, stmt) {
			votedFor = append(votedFor, stmt)
			update = true
		}

		if !canVote(stmt, accepted) {
			// maybe stuck?
			Log("send", "stuck")
			continue
		}

		if !inArray(nodes, n.ID) { // we are not in these nodes
			continue
		}

		if n.checkQuorum(nodes) {
			if !inArray(confirmed, stmt) {
				Log("put", stmt+" end")
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
			VotedFor:     votedFor,
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

	stmt := in.Key + "=" + in.Value
	Log("put", stmt+" start")
	votedFor := n.NodesState[n.ID].VotedFor
	if !canVote(stmt, votedFor) {
		return &kv.PutResponse{Ret: kv.ReturnCode_FAILURE}, nil
	}

	n.StateCounter++
	n.NodesState[n.ID] = fvp.SendMsg_State{
		Id:           n.ID,
		Accepted:     n.NodesState[n.ID].Accepted,
		VotedFor:     append(votedFor, stmt),
		Confirmed:    n.NodesState[n.ID].Confirmed,
		QuorumSlices: n.NodesState[n.ID].QuorumSlices,
		Counter:      n.StateCounter,
	}

	return &kv.PutResponse{Ret: kv.ReturnCode_SUCCESS}, nil
}

func (n *node) createNode() {

	n.NodesAddrs = make([]string, 0)
	n.NodesState = make(map[string]fvp.SendMsg_State, 0)
	n.NodesQuorumSlices = make(map[string][][]string, 0)
	n.NodesFvpClients = make(map[string]fvp.ServerClient, 0)

	// Build our own slices
	ourSlices := make([]*fvp.SendMsg_Slice, 0)
	for _, slice := range n.QsSlices {
		for _, node := range slice {
			if !inArray(n.NodesAddrs, node) {
				n.NodesAddrs = append(n.NodesAddrs, node)
			}
		}
		s := &fvp.SendMsg_Slice{
			Nodes: slice,
		}
		ourSlices = append(ourSlices, s)
	}

	// set our entry in NodesQuorumSlices
	n.NodesQuorumSlices[n.ID] = n.QsSlices

	// append our own state to NodesState
	ourState := fvp.SendMsg_State{
		Accepted:     make([]string, 0),
		Confirmed:    make([]string, 0),
		VotedFor:     make([]string, 0),
		Counter:      0,
		Id:           n.ID,
		QuorumSlices: ourSlices,
	}
	n.NodesState[n.ID] = ourState

	n.Dictionary = make(map[string]string, 0)
}

func (n *node) buildClients() {
	// grpc will retry in 20 ms at most 5 times when failed
	opts := []grpc_retry.CallOption{
		grpc_retry.WithMax(5),
		grpc_retry.WithPerRetryTimeout(timeout),
	}

	for _, addr := range n.NodesAddrs {

		Log("connection", "Connecting to "+addr)

		conn, err := grpc.Dial(addr, grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)))
		if err != nil {
			Log("connection", fmt.Sprintf("Failed to connect to %s. %v\n", addr, err))
		}

		n.NodesFvpClients[addr] = fvp.NewServerClient(conn)
	}
}

var (
	n          node
	configFile = flag.String("config", "cfg.json", "the file to read the configuration from")
	print      = flag.Bool("print", false, "to print the state")
	help       = flag.Bool("h", false, "for usage")
	timeout    = 100 * time.Millisecond
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

	// setup grpc
	lis, err := net.Listen("tcp", ":"+strings.Split(n.ID, ":")[1])
	if err != nil {
		Log("connection", fmt.Sprintf("Failed to listen on the port. %v", err))
	}

	grpcServer := grpc.NewServer()
	fvp.RegisterServerServer(grpcServer, &n)
	kv.RegisterKeyValueStoreServer(grpcServer, &n)
	n.buildClients()

	Log("connection", "Listening on "+n.ID)

	ticker := time.NewTicker(500 * time.Millisecond)
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
