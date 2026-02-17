package merkle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var (
	ErrMissingSegment = errors.New("merkle.segment.missing")
	ErrDivergentRoot  = errors.New("merkle.root.divergent")
	ErrInvalidProof   = errors.New("merkle.proof.invalid")
)

type Builder struct {
	leaves [][]byte
}

func NewBuilder() *Builder {
	return &Builder{leaves: make([][]byte, 0)}
}

func (b *Builder) AddLeaf(data []byte) {
	b.leaves = append(b.leaves, hashLeaf(data))
}

func (b *Builder) Root() string {
	if len(b.leaves) == 0 {
		return ""
	}
	layer := b.leaves
	for len(layer) > 1 {
		layer = buildLayer(layer)
	}
	return hex.EncodeToString(layer[0])
}

func (b *Builder) Proof(index int) ([]string, error) {
	if index < 0 || index >= len(b.leaves) {
		return nil, ErrMissingSegment
	}
	proof := make([]string, 0)
	layer := b.leaves
	idx := index
	for len(layer) > 1 {
		siblingIdx := idx ^ 1
		if siblingIdx < len(layer) {
			proof = append(proof, hex.EncodeToString(layer[siblingIdx]))
		}
		layer = buildLayer(layer)
		idx /= 2
	}
	return proof, nil
}

func VerifyProof(root string, data []byte, proof []string, index int, totalLeaves int) error {
	if totalLeaves <= 0 || index < 0 || index >= totalLeaves {
		return ErrInvalidProof
	}
	calculated := hashLeaf(data)
	pos := index
	for _, sibling := range proof {
		siblingBytes, err := hex.DecodeString(sibling)
		if err != nil {
			return ErrInvalidProof
		}
		if pos%2 == 0 {
			calculated = hashPair(calculated, siblingBytes)
		} else {
			calculated = hashPair(siblingBytes, calculated)
		}
		pos /= 2
	}
	expectedRootBytes, err := hex.DecodeString(root)
	if err != nil {
		return ErrInvalidProof
	}
	if !bytes.Equal(calculated, expectedRootBytes) {
		return ErrDivergentRoot
	}
	return nil
}

func buildLayer(previous [][]byte) [][]byte {
	next := make([][]byte, 0, (len(previous)+1)/2)
	for i := 0; i < len(previous); i += 2 {
		if i+1 >= len(previous) {
			next = append(next, previous[i])
			continue
		}
		next = append(next, hashPair(previous[i], previous[i+1]))
	}
	return next
}

func hashLeaf(data []byte) []byte {
	h := sha256.Sum256(append([]byte{0x00}, data...))
	return h[:]
}

func hashPair(left, right []byte) []byte {
	h := sha256.Sum256(append(append([]byte{0x01}, left...), right...))
	return h[:]
}
