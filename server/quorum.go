package main

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
