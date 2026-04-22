package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// Domain separation prefixes per RFC 6962.
const (
	LeafSeparator     = 0x00
	InternalSeparator = 0x01
)

var (
	ErrEmptyTree       = errors.New("tree is empty")
	ErrIndexOutOfBounds = errors.New("index out of bounds")
)

// Hash represents a 32-byte SHA-256 digest.
type Hash [32]HashValue

type HashValue = byte

// Hex returns the hexadecimal representation of the hash.
func (h Hash) Hex() string {
	return hex.EncodeToString(h[:])
}

// HashFromHex parses a hex string into a Hash.
func HashFromHex(s string) Hash {
	var h Hash
	b, err := hex.DecodeString(s)
	if err == nil && len(b) == 32 {
		copy(h[:], b)
	}
	return h
}

// Tree represents an append-only Merkle tree.
type Tree struct {
	leaves []Hash
}

// NewTree creates a new empty Merkle tree.
func NewTree() *Tree {
	return &Tree{
		leaves: make([]Hash, 0),
	}
}

// LeafHash computes the SHA-256(0x00 || data) for a leaf.
func LeafHash(data []byte) Hash {
	h := sha256.New()
	h.Write([]byte{LeafSeparator})
	h.Write(data)
	var out Hash
	copy(out[:], h.Sum(nil))
	return out
}

// NodeHash computes the SHA-256(0x01 || left || right) for an internal node.
func NodeHash(left, right Hash) Hash {
	h := sha256.New()
	h.Write([]byte{InternalSeparator})
	h.Write(left[:])
	h.Write(right[:])
	var out Hash
	copy(out[:], h.Sum(nil))
	return out
}

// Append adds a new leaf to the tree.
func (t *Tree) Append(leaf Hash) {
	t.leaves = append(t.leaves, leaf)
}

// Size returns the number of leaves in the tree.
func (t *Tree) Size() int {
	return len(t.leaves)
}

// Root returns the Merkle Tree Root for the current size.
// Returns an empty hash if the tree is empty.
func (t *Tree) Root() Hash {
	if len(t.leaves) == 0 {
		return Hash{}
	}
	return t.computeRoot(t.leaves)
}

func (t *Tree) computeRoot(leaves []Hash) Hash {
	n := len(leaves)
	if n == 0 {
		return Hash{}
	}
	if n == 1 {
		return leaves[0]
	}

	// Largest power of 2 less than n
	k := largestPowerOf2LessThan(n)
	left := t.computeRoot(leaves[:k])
	right := t.computeRoot(leaves[k:])
	return NodeHash(left, right)
}

// InclusionProof generates an RFC 6962 inclusion proof for the leaf at index.
func (t *Tree) InclusionProof(index int) ([]Hash, error) {
	if index < 0 || index >= len(t.leaves) {
		return nil, ErrIndexOutOfBounds
	}
	return t.computeInclusionProof(t.leaves, index), nil
}

func (t *Tree) computeInclusionProof(leaves []Hash, m int) []Hash {
	n := len(leaves)
	if n <= 1 {
		return nil
	}

	k := largestPowerOf2LessThan(n)
	if m < k {
		proof := t.computeInclusionProof(leaves[:k], m)
		proof = append(proof, t.computeRoot(leaves[k:]))
		return proof
	} else {
		proof := t.computeInclusionProof(leaves[k:], m-k)
		proof = append(proof, t.computeRoot(leaves[:k]))
		return proof
	}
}

// largestPowerOf2LessThan returns the largest power of 2 less than n.
// n must be > 1.
func largestPowerOf2LessThan(n int) int {
	if n <= 1 {
		return 0
	}
	res := 1
	for res < n {
		res <<= 1
	}
	return res >> 1
}

// VerifyInclusion verifies that leaf is at index in a tree of size with root.
func VerifyInclusion(root Hash, leaf Hash, index int, size int, proof []Hash) bool {
	if index < 0 || index >= size {
		return false
	}
	if size == 1 {
		return len(proof) == 0 && root == leaf
	}

	// We need to know which side each sibling came from.
	// Since the proof is bottom-up but the traversal to find sides is top-down,
	// we first determine the sides by following the path.
	type side struct {
		isRight bool
		hash    Hash
	}
	var path []bool
	fs := size
	fn := index
	for fs > 1 {
		k := largestPowerOf2LessThan(fs)
		if fn < k {
			path = append(path, false) // We went left, sibling is Right
			fs = k
		} else {
			path = append(path, true) // We went right, sibling is Left
			fn -= k
			fs -= k
		}
	}

	if len(proof) != len(path) {
		return false
	}

	current := leaf
	// Process path in reverse (bottom-up)
	for i := len(path) - 1; i >= 0; i-- {
		isRightBranch := path[i]
		sibling := proof[len(path)-1-i]
		if isRightBranch {
			// We were the right child, sibling is left
			current = NodeHash(sibling, current)
		} else {
			// We were the left child, sibling is right
			current = NodeHash(current, sibling)
		}
	}

	return current == root
}
