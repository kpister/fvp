package main

import (
	fvp "github.com/kpister/fvp/server/proto/fvp"
	kv "github.com/kpister/fvp/server/proto/kvstore"

	"context"
	"time"
)

type node struct {
}

func (n *node) broadcast() {
	/*
	   for n' in neighbors {
	       n'.send(n.state)
	   }
	*/
}

func (n *node) Send(ctx context.Context, in *fvp.SendMsg) (*fvp.EmptyMessage, error) {

	// check_blocking(in.votedfor)
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
	return &node{}
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
