// Package ccl provides implementations of Connected Component Labeling algorithms.
package ccl

import (
	"math"
	"sort"
)

const (
	// NullLabel is the label that should be initially given to all nodes.
	NullLabel = -1
)

// A CCLabeler is a container of nodes that can be labeled.
type CCLabeler interface {
	// Reset resets the iterator of the container.
	Reset()

	// Next moves the iterator forward to the next node to be labeled.
	// If there are no more nodes available, Next returns false.
	Next() bool

	// Neighbors returns the labels of the neighbors of the current node.
	Neighbors() []int

	// GetLabel returns the label of the current node.
	GetLabel() int

	// SetLabel sets the label of the current node.
	SetLabel(int)

	// Size returns the size of the current node.
	// The sum of the sizes of all nodes sharing a same label is the size of that label.
	// Labels are assigned in descending order of their sizes.
	Size() int
}

type blob struct {
	label int
	size  int
}

// HoshenKopelman labels all nodes in the container using the Hoshen-Kopelman algorithm.
// It returns the sizes of all labels in descending order.
func HoshenKopelman(labeler CCLabeler) []int {
	// equivalence implements a disjoint-size data structure.
	// equivalence[i] == j, means labels i and j are in the same cluster.
	equivalence := make([]int, 0)
	find := func(label int) int {
		parent := label
		for equivalence[parent] != parent {
			parent = equivalence[parent]
		}

		// Squash equivalence links to speed up future lookups.
		for equivalence[label] != label {
			tmp := equivalence[label]
			equivalence[label] = parent
			label = tmp
		}

		return parent
	}

	largestLabel := 0
	blobs := make([]blob, 0)
	// First pass.
	for labeler.Next() {
		neighbors := labeler.Neighbors()
		// If a node has no neighbors, give it a new label.
		if len(neighbors) == 0 {
			labeler.SetLabel(largestLabel)
			equivalence = append(equivalence, largestLabel)
			blobs = append(blobs, blob{label: largestLabel, size: labeler.Size()})
			largestLabel++
			continue
		}

		// Find the smallest label among neighbors.
		roots := make([]int, 0, len(neighbors))
		min := math.MaxInt
		for _, n := range neighbors {
			root := find(n)
			roots = append(roots, root)
			if root < min {
				min = root
			}
		}

		// Set the root of all clusters to the smallest label,
		// which means now all clusters are of the same label.
		for i, root := range roots {
			equivalence[root] = min
			find(neighbors[i])
		}
		labeler.SetLabel(min)
		blobs[min].size += labeler.Size()
	}

	// Make the list of labels a contiguous sequence of natural numbers.
	for _, b := range blobs {
		root := find(b.label)
		if root == b.label {
			continue
		}
		blobs[root].size += b.size
		blobs[b.label].size = 0
	}
	sort.Slice(blobs, func(i, j int) bool { return blobs[i].size > blobs[j].size })
	mapping := make([]int, len(blobs))
	numNotEmpty := 0
	for i, b := range blobs {
		if b.size == 0 {
			mapping[b.label] = mapping[find(b.label)]
		} else {
			mapping[b.label] = i
			numNotEmpty++
		}
	}
	blobs = blobs[:numNotEmpty]

	// Second pass.
	labeler.Reset()
	for labeler.Next() {
		label := labeler.GetLabel()
		mapped := mapping[label]
		labeler.SetLabel(mapped)
	}

	// Return the sizes of the assigned labels.
	sizes := make([]int, 0, len(blobs))
	for _, b := range blobs {
		sizes = append(sizes, b.size)
	}
	return sizes
}
