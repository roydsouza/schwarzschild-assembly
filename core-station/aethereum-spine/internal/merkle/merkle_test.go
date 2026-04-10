package merkle

import (
	"bytes"
	"testing"
)

func TestMerkleTree_Root(t *testing.T) {
	tree := NewTree()
	
	// Test empty tree
	root := tree.Root()
	if !bytes.Equal(root[:], make([]byte, 32)) {
		t.Errorf("expected empty root, got %s", root.Hex())
	}
	
	// Test 1 leaf
	leaf1 := LeafHash([]byte("hello"))
	tree.Append(leaf1)
	if tree.Root() != leaf1 {
		t.Errorf("expected root to be leaf1, got %s", tree.Root().Hex())
	}
	
	// Test 2 leaves
	leaf2 := LeafHash([]byte("world"))
	tree.Append(leaf2)
	expectedRoot2 := NodeHash(leaf1, leaf2)
	if tree.Root() != expectedRoot2 {
		t.Errorf("expected root to be NodeHash(leaf1, leaf2), got %s", tree.Root().Hex())
	}
	
	// Test 3 leaves
	leaf3 := LeafHash([]byte("!"))
	tree.Append(leaf3)
	// Root(3) = NodeHash(Root(2), Root(1))
	expectedRoot3 := NodeHash(expectedRoot2, leaf3)
	if tree.Root() != expectedRoot3 {
		t.Errorf("expected root 3, got %s", tree.Root().Hex())
	}
	
	// Test 4 leaves
	leaf4 := LeafHash([]byte("?"))
	tree.Append(leaf4)
	// Root(4) = NodeHash(Root(2), NodeHash(leaf3, leaf4))
	expectedRoot4 := NodeHash(expectedRoot2, NodeHash(leaf3, leaf4))
	if tree.Root() != expectedRoot4 {
		t.Errorf("expected root 4, got %s", tree.Root().Hex())
	}
}

func TestMerkleTree_InclusionProof(t *testing.T) {
	tree := NewTree()
	data := [][]byte{
		[]byte("apple"),
		[]byte("banana"),
		[]byte("cherry"),
		[]byte("date"),
		[]byte("elderberry"),
	}
	
	for _, d := range data {
		tree.Append(LeafHash(d))
	}
	
	root := tree.Root()
	size := tree.Size()
	
	for i, d := range data {
		leaf := LeafHash(d)
		proof, err := tree.InclusionProof(i)
		if err != nil {
			t.Fatalf("failed to get proof for index %d: %v", i, err)
		}
		
		if !VerifyInclusion(root, leaf, i, size, proof) {
			t.Errorf("failed to verify inclusion for index %d", i)
		}
	}
	
	// Test invalid index
	_, err := tree.InclusionProof(size)
	if err == nil {
		t.Error("expected error for out of bounds index")
	}
}

func TestLargestPowerOf2(t *testing.T) {
	tests := []struct {
		n    int
		want int
	}{
		{2, 1},
		{3, 2},
		{4, 2},
		{5, 4},
		{7, 4},
		{8, 4},
		{9, 8},
		{15, 8},
		{16, 8},
	}
	
	for _, tt := range tests {
		if got := largestPowerOf2LessThan(tt.n); got != tt.want {
			t.Errorf("largestPowerOf2LessThan(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}
