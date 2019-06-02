package main

import (
	"encoding/json"
	"fmt"
	"math/rand"

	fvp "github.com/kpister/fvp/server/proto/fvp"
)

/*
Returns true if the message should be dropped otherwise returns false
*/
// func (n *node) dropMessageChaos(from int32) bool {
// 	if from < 0 || from >= int32(len(n.ServersAddr)) || n.ID < 0 || n.ID >= int32(len(n.ServersAddr)) {
// 		return false
// 	}

// 	random0to1 := rand.Float32()
// 	// n.Chaos is the probability with which we want to drop the value
// 	// 0 - no drop 1 - drop every message
// 	if random0to1 > n.Chaos[from][n.ID] {
// 		return false
// 	}
// 	return true
// }

func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func inArray(arr []string, checkEle string) bool {
	for _, e := range arr {
		if e == checkEle {
			return true
		}
	}
	return false
}

func remove(s [][]string, i int) [][]string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func isUnanimousVote(slice []string, statement string, state map[string]fvp.SendMsg_State) bool {
	for _, node := range slice {
		if !inArray(state[node].VotedFor, statement) {
			return false
		}
	}
	return true
}

func prettyPrintMap(inmap interface{}) {
	mapB, _ := json.MarshalIndent(inmap, "", " ")
	fmt.Println("MAP::", string(mapB))
}

func getVote(prob float32, option1 string, option2 string) string {
	random0to1 := rand.Float32()
	fmt.Println(random0to1)
	if random0to1 < prob {
		return option1
	}
	return option2
}

// func resizeSlice(a []*raft.Entry, newSize int) []*raft.Entry {
// 	return append([]*raft.Entry(nil), a[:newSize]...)
// }
