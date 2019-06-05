package main

import (
	"encoding/json"
	"flag"
	"fmt"
	fvp "github.com/kpister/fvp/server/proto/fvp"
	"google.golang.org/grpc"
	"io/ioutil"
	"os"
)

func (n *node) buildClients() {

	for _, addr := range n.NodesAddrs {

		Log("connection", "Connecting to "+addr)

		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			Log("connection", fmt.Sprintf("Failed to connect to %s. %v\n", addr, err))
		}

		n.NodesFvpClients[addr] = fvp.NewServerClient(conn)
	}
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

func init() {
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
