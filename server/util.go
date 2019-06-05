package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

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

func isUnanimous(slice []string, statement string, state map[string]fvp.SendMsg_State, msgtype string) bool {
	for _, node := range slice {
		_, ok := state[node]
		// fmt.Println(node, "in state?", ok)
		if msgtype == "vote" {
			if !ok || !inArray(state[node].VotedFor, statement) {
				return false
			}
		} else {
			if !ok || !inArray(state[node].Accepted, statement) {
				return false
			}
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
	if random0to1 < prob {
		return option1
	}
	return option2
}

func getVoteTC3(id string) string {
	if id == "localhost:8001" || id == "localhost:8002" {
		return "a=1"
	}
	return "a=2" // for localhost:8004 and :8005

}

func getVoteTC4(id string) string {
	if id == "localhost:8001" || id == "localhost:8002" {
		return "a=1"
	} else if id == "localhost:8005" || id == "localhost:8006" {
		return "a=2"
	}
	return "b=1" // for localhost:8004

}

func convertQuorumSlices(qs []*fvp.SendMsg_Slice) [][]string {
	ret := make([][]string, 0)

	for _, el := range qs {
		ret = append(ret, el.Nodes)
	}
	return ret
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

// func resizeSlice(a []*raft.Entry, newSize int) []*raft.Entry {
// 	return append([]*raft.Entry(nil), a[:newSize]...)
// }
