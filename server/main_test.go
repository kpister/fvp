package main

import (
	"fmt"
	"testing"
)

func TestCheckQuorum(t *testing.T) {
	var n *node
	var nodes []string

	// TODO: could read these settings from testing config files?
	n = createNode("0")
	n.nodesQuorumSlices = map[string][][]string{
		"0": [][]string{[]string{"0", "1", "2"}},
		"1": [][]string{[]string{"0", "1", "2"}},
		"2": [][]string{[]string{"0", "1", "2"}},
	}
	nodes = []string{"0", "1", "2"}
	t.Run("3 mutual connected", testCheckQuorumFunc(n, nodes, true))

	n = createNode("0")
	n.nodesQuorumSlices = map[string][][]string{
		"0": [][]string{[]string{"0", "1", "2"}},
		"1": [][]string{[]string{"1", "0", "2"}},
		"2": [][]string{[]string{"2", "0", "1"}},
		"3": [][]string{[]string{"3", "4", "5"}},
		"4": [][]string{[]string{"4", "3"}},
		"5": [][]string{[]string{"5", "3", "4"}},
	}
	nodes = []string{"0", "1", "2", "3", "4", "5"}
	t.Run("2 separated groups", testCheckQuorumFunc(n, nodes, true))

	n = createNode("1")
	n.nodesQuorumSlices = map[string][][]string{
		"1": [][]string{[]string{"1", "2", "3"}},
		"2": [][]string{[]string{"2", "3", "4"}},
		"3": [][]string{[]string{"2", "3", "4"}},
		"4": [][]string{[]string{"2", "3", "4"}},
	}
	nodes = []string{"1", "2", "3"}
	t.Run("example on page 5 false", testCheckQuorumFunc(n, nodes, false))

	nodes = []string{"1", "2", "3", "4"}
	t.Run("example on page 5 true", testCheckQuorumFunc(n, nodes, true))

	n = createNode("1")
	n.nodesQuorumSlices = map[string][][]string{
		"1": [][]string{[]string{"1", "2"}},
		"2": [][]string{[]string{"2", "3"}},
		"3": [][]string{[]string{"3", "4"}},
		"4": [][]string{[]string{"4", "5"}},
		"5": [][]string{[]string{"5", "6"}},
		"6": [][]string{[]string{"5", "1"}},
	}
	nodes = []string{"1", "2", "3", "4", "5", "6"}
	t.Run("example on page 6 true", testCheckQuorumFunc(n, nodes, true))

	nodes = []string{"1", "2", "4", "5", "6"}
	t.Run("example on page 6 false", testCheckQuorumFunc(n, nodes, false))
}

func testCheckQuorumFunc(n *node, nodes []string, expected bool) func(*testing.T) {
	return func(t *testing.T) {
		isQuorum := n.checkQuorum(nodes)
		if isQuorum != expected {
			t.Error(fmt.Sprintf("Expected %v but got %v", expected, isQuorum))
		}
	}
}

func TestCheckBlocking(t *testing.T) {
	var n *node
	var nodes []string

	n = createNode("0")
	n.nodesQuorumSlices = map[string][][]string{
		"0": [][]string{[]string{"0", "1", "2"}, {"0", "3", "4"}, {"0", "4", "5"}},
	}
	nodes = []string{"1", "3", "5"}
	t.Run("example in blog post 1", testCheckBlockingFunc(n, nodes, true))

	nodes = []string{"1", "4"}
	t.Run("example in blog post 2", testCheckBlockingFunc(n, nodes, true))

	nodes = []string{"1", "5"}
	t.Run("example in blog post 3", testCheckBlockingFunc(n, nodes, false))
}

func testCheckBlockingFunc(n *node, nodes []string, expected bool) func(*testing.T) {
	return func(t *testing.T) {
		isBlocking := n.checkBlocking(nodes)
		if isBlocking != expected {
			t.Error(fmt.Sprintf("Expected %v but got %v", expected, isBlocking))
		}
	}
}
