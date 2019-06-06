package main

import (
	"context"
	fvp "github.com/kpister/fvp/server/proto/fvp"
	kv "github.com/kpister/fvp/server/proto/kvstore"
)

func (n *node) IncrementTerm(ctx context.Context, in *kv.EmptyMessage) (*kv.EmptyMessage, error) {
	n.Term++
	ourState := fvp.SendMsg_State{
		Accepted:     make([]string, 0),
		Confirmed:    make([]string, 0),
		VotedFor:     make([]string, 0),
		Counter:      0,
		Id:           n.ID,
		QuorumSlices: n.NodesState[n.ID].QuorumSlices,
	}

	n.NodesState[n.ID] = ourState

	return &kv.EmptyMessage{}, nil
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
