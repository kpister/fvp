package main

import (
	"context"

	fvp "github.com/kpister/fvp/server/proto/fvp"
	kv "github.com/kpister/fvp/server/proto/kvstore"
)

func (n *node) IncrementTerm(ctx context.Context, in *kv.EmptyMessage) (*kv.EmptyMessage, error) {
	gl.Lock()

	n.Term++
	ourState := &fvp.SendMsg_State{
		Accepted:     make([]string, 0),
		Confirmed:    make([]string, 0),
		VotedFor:     make([]string, 0),
		Counter:      0,
		Id:           n.ID,
		QuorumSlices: n.NodesState[n.ID].QuorumSlices,
	}

	n.NodesState = make(map[string]*fvp.SendMsg_State)
	n.NodesState[n.ID] = ourState

	gl.Unlock()

	return &kv.EmptyMessage{}, nil
}

func (n *node) Get(ctx context.Context, in *kv.GetRequest) (*kv.GetResponse, error) {
	// lookup in.key in dictionary
	gl.Lock()
	val, ok := n.Dictionary[in.Key]
	gl.Unlock()

	if ok {
		return &kv.GetResponse{Value: val, Ret: kv.ReturnCode_SUCCESS}, nil
	}

	return &kv.GetResponse{Value: "", Ret: kv.ReturnCode_FAILURE}, nil
}

func (n *node) Put(ctx context.Context, in *kv.PutRequest) (*kv.PutResponse, error) {
	// set value in dictionary
	gl.Lock()
	defer gl.Unlock()

	stmt := in.Key + "=" + in.Value
	if !canVote(stmt, n.NodesState[n.ID].VotedFor) {
		return &kv.PutResponse{Ret: kv.ReturnCode_FAILURE}, nil
	}
	Log(n.Term, "put", stmt+" start")

	n.StateCounter++
	n.NodesState[n.ID].VotedFor = append(n.NodesState[n.ID].VotedFor, stmt)

	return &kv.PutResponse{Ret: kv.ReturnCode_SUCCESS}, nil
}
