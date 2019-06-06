package main

import (
	"context"
	fvp "github.com/kpister/fvp/server/proto/fvp"
	"math/rand"
	"strconv"
	"strings"
)

func (n *node) evilBehavior(strategy string) {
	// all good ndoes are doing a=1, we do a=2

	for _, addr := range n.NodesAddrs {
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
			randVotedFor := make([]string, 0)
			randAccepted := make([]string, 0)
			for _, stmt := range n.NodesState[n.ID].VotedFor {
				key := strings.Split(stmt, "=")[0]
				value := rand.Int() % 10
				randVotedFor = append(randVotedFor, key+"="+strconv.Itoa(value))
			}
			for _, stmt := range n.NodesState[n.ID].Accepted {
				key := strings.Split(stmt, "=")[0]
				value := rand.Int() % 10
				randAccepted = append(randAccepted, key+"="+strconv.Itoa(value))
			}

			fakestate = fvp.SendMsg_State{
				Id:           n.ID,
				Accepted:     randAccepted,
				VotedFor:     randVotedFor,
				Confirmed:    []string{},
				QuorumSlices: n.NodesState[n.ID].QuorumSlices,
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
			Log(n.Term, "broadcast", "Invalid evil strategy")
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
		ctx := context.Background()

		_, err := n.NodesFvpClients[addr].Send(ctx, args)
		if err != nil {
			n.errorHandler(err, "broadcast", addr)
		}

	}

}
