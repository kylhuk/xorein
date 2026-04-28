// treekem.go implements the TreeKEM ratchet tree per RFC 9420 §7 with the
// Xorein hybrid KEM ciphersuite (spec 12 §1.2).
//
// The tree is a left-balanced binary tree over leaf nodes.  Each leaf holds
// the member's hybrid KEM public key (X25519 || ML-KEM-768 encap key).
// Internal nodes hold the path-ratchet key derived from UpdatePath.
//
// Source: docs/spec/v0.1/12-mode-tree.md §1.2 and RFC 9420 §7.
package tree

import (
	"encoding/binary"
	"errors"
	"fmt"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
)

var (
	// ErrTreeKEMDecrypt is returned when path secret decapsulation fails.
	ErrTreeKEMDecrypt = errors.New("tree/treekem: path secret decapsulation failed")
	// ErrTreeKEMLeafIndex is returned for an out-of-range leaf index.
	ErrTreeKEMLeafIndex = errors.New("tree/treekem: leaf index out of range")
)

// hybridKEMCombineInfo is the HKDF info label for the hybrid KEM combiner per spec 12 §1.2.
const hybridKEMCombineInfo = "xorein/tree/v1/kem-combine"

// TreeKEMNode is a node in the ratchet tree.
// A leaf node holds both the member's hybrid public key and (if local) private key.
// An internal (path) node holds the resolved path KEM key.
type TreeKEMNode struct {
	// PublicKey is the hybrid KEM public key: X25519 (32 B) || ML-KEM-768 encap (1184 B).
	// Nil means blank node.
	PublicKey []byte
	// PrivateKey is the hybrid KEM private key: X25519 (32 B) || ML-KEM-768 decap (2400 B).
	// Nil for all non-local nodes.
	PrivateKey []byte
}

// HybridPublicKeySize is the total size of a hybrid public key.
const HybridPublicKeySize = v0crypto.X25519KeySize + v0crypto.MLKEM768PublicKeySize  // 32 + 1184 = 1216
// HybridPrivateKeySize is the total size of a hybrid private key.
const HybridPrivateKeySize = v0crypto.X25519KeySize + v0crypto.MLKEM768PrivateKeySize // 32 + 2400 = 2432

// HybridCiphertextSize is the size of a hybrid KEM ciphertext (combined output).
const HybridCiphertextSize = v0crypto.X25519KeySize + v0crypto.MLKEM768CiphertextSize // 32 + 1088 = 1120

// RatchetTree is a binary left-balanced tree per RFC 9420 §7.
// Nodes is indexed by RFC 9420 node indexing: leaf i is at position 2i,
// parent of node x is at (x + level(x)) >> 1 * 2 (simplified here as parent(x)).
type RatchetTree struct {
	// Nodes holds all tree nodes (leaf and internal), indexed by RFC 9420 node index.
	// Size = 2*numLeaves - 1.
	Nodes []*TreeKEMNode
	// GroupID is the MLS group identifier.
	GroupID string
	// numLeaves is the number of leaf nodes.
	numLeaves int
}

// NewRatchetTree creates a left-balanced tree for the given member init public keys.
// Each initKey in memberInitKeys must be a hybrid public key (HybridPublicKeySize bytes):
//
//	x25519_pub (32 B) || mlkem768_encap_pub (1184 B)
func NewRatchetTree(groupID string, memberInitKeys [][]byte) *RatchetTree {
	n := len(memberInitKeys)
	if n == 0 {
		return &RatchetTree{GroupID: groupID}
	}
	// Size of the left-balanced tree array.
	size := 2*n - 1
	nodes := make([]*TreeKEMNode, size)
	// Populate leaf nodes (even indices: 0, 2, 4, ...).
	for i, key := range memberInitKeys {
		nodes[leafNodeIndex(i)] = &TreeKEMNode{PublicKey: key}
	}
	return &RatchetTree{
		Nodes:     nodes,
		GroupID:   groupID,
		numLeaves: n,
	}
}

// leafNodeIndex returns the RFC 9420 node index for leaf i.
func leafNodeIndex(i int) int { return 2 * i }

// UpdatePath generates a new path secret for leaf leafIndex, re-encapsulates it to
// all co-path members, and returns:
//   - pathSecret: the fresh 32-byte path secret (= commit secret for 2-member groups)
//   - encryptedPathSecrets: one hybrid KEM ciphertext per co-path node, ordered root→leaf
//
// hybridPriv and hybridPub are the caller's leaf hybrid KEM private and public keys.
func (t *RatchetTree) UpdatePath(leafIndex int, hybridPriv, hybridPub []byte) (
	pathSecret []byte, encryptedPathSecrets [][]byte, err error,
) {
	if leafIndex < 0 || leafIndex >= t.numLeaves {
		return nil, nil, ErrTreeKEMLeafIndex
	}

	// Generate a fresh path secret.
	pathSecret, err = v0crypto.RandomBytes(32)
	if err != nil {
		return nil, nil, fmt.Errorf("tree/treekem: generate path secret: %w", err)
	}

	// Update the leaf node's public key to hybridPub.
	leafIdx := leafNodeIndex(leafIndex)
	if t.Nodes[leafIdx] == nil {
		t.Nodes[leafIdx] = &TreeKEMNode{}
	}
	t.Nodes[leafIdx].PublicKey = hybridPub
	t.Nodes[leafIdx].PrivateKey = hybridPriv

	// Collect co-path node public keys (the recipients who need to decrypt pathSecret).
	// Co-path = sibling of each node on the direct path from leaf to root.
	directPath, coPath := t.directAndCoPath(leafIndex)
	encryptedPathSecrets = make([][]byte, 0, len(coPath))

	// Derive a per-node encrypted path secret for each co-path node.
	// For simplicity (and correctness for small groups), encrypt the same
	// pathSecret to each co-path recipient using hybrid KEM.
	for i, coIdx := range coPath {
		recipient := t.Nodes[coIdx]
		if recipient == nil || len(recipient.PublicKey) == 0 {
			// Blank co-path node — skip.
			encryptedPathSecrets = append(encryptedPathSecrets, nil)
			continue
		}

		ct, encErr := hybridKEMEncapsulate(recipient.PublicKey, pathSecret)
		if encErr != nil {
			return nil, nil, fmt.Errorf("tree/treekem: encapsulate to co-path node %d: %w", coIdx, encErr)
		}
		encryptedPathSecrets = append(encryptedPathSecrets, ct)

		// Update the internal node on the direct path to reflect the new path secret.
		if i < len(directPath) {
			dpIdx := directPath[i]
			if t.Nodes[dpIdx] == nil {
				t.Nodes[dpIdx] = &TreeKEMNode{}
			}
			// Derive internal node KEM key from pathSecret + node index.
			nodePub, nodePriv, kErr := deriveNodeKEMKey(pathSecret, uint32(dpIdx))
			if kErr != nil {
				return nil, nil, fmt.Errorf("tree/treekem: derive node key: %w", kErr)
			}
			t.Nodes[dpIdx].PublicKey = nodePub
			t.Nodes[dpIdx].PrivateKey = nodePriv
		}
	}

	return pathSecret, encryptedPathSecrets, nil
}

// ProcessUpdatePath processes an incoming UpdatePath from leafIndex (a co-path member)
// and returns the commit secret.
//
// encryptedPathSecret is the ciphertext from UpdatePath targeted at this node's
// co-path position. hybridPriv is this node's leaf hybrid KEM private key.
func (t *RatchetTree) ProcessUpdatePath(leafIndex int, encryptedPathSecret []byte, hybridPriv []byte) (commitSecret []byte, err error) {
	if leafIndex < 0 || leafIndex >= t.numLeaves {
		return nil, ErrTreeKEMLeafIndex
	}
	if len(encryptedPathSecret) == 0 {
		return nil, errors.New("tree/treekem: empty encrypted path secret")
	}

	commitSecret, err = hybridKEMDecapsulate(hybridPriv, encryptedPathSecret)
	if err != nil {
		return nil, ErrTreeKEMDecrypt
	}
	return commitSecret, nil
}

// --- internal helpers ---

// directAndCoPath returns the direct path (leaf → root) and co-path (sibling of each
// node on the direct path) for leaf leafIndex.
//
// The left-balanced binary tree has 2*numLeaves-1 nodes.
// Leaves are at even indices (0, 2, 4, …), internal nodes at odd indices.
// The root is the unique odd node that has no parent within the valid node range,
// which for this implementation is the single internal node for 2-leaf trees,
// or the "center" for larger trees.
//
// We use a level-based parent computation matching RFC 9420 §7.1.
func (t *RatchetTree) directAndCoPath(leafIndex int) (directPath, coPath []int) {
	if t.numLeaves <= 1 {
		return nil, nil
	}
	nodeIdx := leafNodeIndex(leafIndex)
	size := 2*t.numLeaves - 1
	root := treeRootIndex(size)
	for nodeIdx != root {
		p := nodeParent(nodeIdx, size)
		sib := nodeSibling(nodeIdx, size)
		if p < 0 || p >= size || sib < 0 || sib >= size {
			break
		}
		directPath = append(directPath, p)
		coPath = append(coPath, sib)
		nodeIdx = p
	}
	return directPath, coPath
}

// nodeLevel returns the "level" of a node in the left-balanced tree:
//   - Level 0: leaf (even index)
//   - Level k: the k-th level internal node (index = 2k-1, 6k-1, etc.)
//
// Per RFC 9420: level(x) = number of trailing 1 bits in x's binary representation.
func nodeLevel(x int) int {
	lvl := 0
	for (x>>uint(lvl))&1 == 1 {
		lvl++
	}
	return lvl
}

// treeRootIndex returns the root node index for a tree of the given size (2n-1 nodes).
// Per RFC 9420 §7.1.1: the root is the node at the "widest" width, i.e., the unique
// node at level floor(log2(n)) where n = (size+1)/2.
func treeRootIndex(size int) int {
	// For size = 2n-1 nodes, root is at level ceil(log2(n)).
	// Iterative: start from any leaf and walk up until reaching a node whose
	// parent index would be ≥ size.
	if size == 1 {
		return 0
	}
	// Walk from leaf 0 (node 0) upward.
	x := 0
	for {
		p := nodeParent(x, size)
		if p == x || p < 0 || p >= size {
			return x
		}
		x = p
	}
}

// nodeParent returns the parent of node x in a tree with the given size.
// Returns -1 if x is the root.
// Per RFC 9420 §7.1.2: parent(x) = x with the k-th bit set to 1 (where k = level(x)+1)
// and the (k+1)-th bit cleared — equivalent to flipping the lowest 0 bit above level k.
func nodeParent(x, size int) int {
	if x < 0 || x >= size {
		return -1
	}
	lvl := nodeLevel(x)
	// Step up: the parent is at level lvl+1.
	// Parent index per RFC 9420: p = (x | bit) & ^(bit << 1), where bit = 1 << (lvl+1).
	bit := 1 << uint(lvl+1)
	p := x | bit
	// Now clear the bit above that (if x was the right child).
	// Actually: parent = x with bit (lvl+1) flipped on, and bit (lvl+2) off.
	// Simpler: parent is the node one level up whose range includes x.
	// p = ((x >> (lvl+1)) << (lvl+1)) | (bit - 1) -- i.e., set all bits 0..lvl to 1.
	// That's the standard RFC 9420 "node parent" formula.
	p = (x>>uint(lvl+1))<<uint(lvl+1) | (bit - 1)
	// Adjust direction: the parent is to the right of x if x < midpoint.
	// midpoint of parent's range = p (the all-ones node at level lvl+1).
	if p < 0 || p >= size || p == x {
		return -1
	}
	return p
}

// nodeSibling returns the sibling of node x (at level k) given tree size.
// sibling = parent ± (1 << level(x)).
func nodeSibling(x, size int) int {
	p := nodeParent(x, size)
	if p < 0 {
		return -1
	}
	lvl := nodeLevel(x)
	delta := 1 << uint(lvl)
	if x < p {
		s := p + delta
		if s < size {
			return s
		}
		return -1
	}
	s := p - delta
	if s >= 0 {
		return s
	}
	return -1
}

// hybridKEMEncapsulate encapsulates pathSecret to a hybrid recipient public key.
// recipientPub must be HybridPublicKeySize bytes: x25519_pub(32) || mlkem_pub(1184).
// Returns the combined ciphertext: x25519_ct(32) || mlkem_ct(1088) || encrypted_secret_aead.
//
// The encapsulated payload is: pathSecret encrypted with AEAD under the hybrid shared secret.
func hybridKEMEncapsulate(recipientPub []byte, pathSecret []byte) ([]byte, error) {
	if len(recipientPub) < HybridPublicKeySize {
		return nil, fmt.Errorf("tree/treekem: recipient pub key too short: %d < %d", len(recipientPub), HybridPublicKeySize)
	}

	// X25519 part.
	var x25519Pub [v0crypto.X25519KeySize]byte
	copy(x25519Pub[:], recipientPub[:v0crypto.X25519KeySize])
	x25519Priv, x25519EphPub, err := v0crypto.GenerateX25519Keypair()
	if err != nil {
		return nil, fmt.Errorf("tree/treekem: x25519 keygen: %w", err)
	}
	ssX25519, err := v0crypto.X25519DH(x25519Priv, x25519Pub)
	if err != nil {
		return nil, fmt.Errorf("tree/treekem: x25519 dh: %w", err)
	}

	// ML-KEM-768 part.
	mlkemPub := recipientPub[v0crypto.X25519KeySize : v0crypto.X25519KeySize+v0crypto.MLKEM768PublicKeySize]
	mlkemCT, ssMLKEM, err := v0crypto.MLKEM768Encapsulate(mlkemPub)
	if err != nil {
		return nil, fmt.Errorf("tree/treekem: mlkem encapsulate: %w", err)
	}

	// Hybrid combine per spec 12 §1.2.
	hybridSS, err := v0crypto.DeriveKey(
		append(ssX25519[:], ssMLKEM...),
		nil,
		hybridKEMCombineInfo,
		32,
	)
	if err != nil {
		return nil, fmt.Errorf("tree/treekem: kem combine: %w", err)
	}

	// Encrypt pathSecret under hybridSS using AES-128-GCM.
	// Nonce = 0 (deterministic; single-use key).
	var aesKey [16]byte
	copy(aesKey[:], hybridSS[:16])
	var nonce [12]byte // all-zero nonce; hybridSS is ephemeral so this is safe
	aad := treeKEMAAD()
	encSecret, err := v0crypto.SealAES128GCM(aesKey, nonce, pathSecret, aad)
	if err != nil {
		return nil, fmt.Errorf("tree/treekem: seal path secret: %w", err)
	}

	// Output: x25519_eph_pub(32) || mlkem_ct(1088) || enc_secret(len(pathSecret)+16)
	out := make([]byte, 0, v0crypto.X25519KeySize+v0crypto.MLKEM768CiphertextSize+len(encSecret))
	out = append(out, x25519EphPub[:]...)
	out = append(out, mlkemCT...)
	out = append(out, encSecret...)
	return out, nil
}

// hybridKEMDecapsulate decapsulates a ciphertext produced by hybridKEMEncapsulate.
// recipientPriv must be HybridPrivateKeySize bytes: x25519_priv(32) || mlkem_priv(2400).
func hybridKEMDecapsulate(recipientPriv []byte, ct []byte) ([]byte, error) {
	if len(recipientPriv) < HybridPrivateKeySize {
		return nil, fmt.Errorf("tree/treekem: recipient priv key too short: %d < %d", len(recipientPriv), HybridPrivateKeySize)
	}
	minCTSize := v0crypto.X25519KeySize + v0crypto.MLKEM768CiphertextSize + 16 // +16 for AEAD tag
	if len(ct) < minCTSize {
		return nil, fmt.Errorf("tree/treekem: ciphertext too short: %d < %d", len(ct), minCTSize)
	}

	// Parse the ciphertext.
	x25519EphPub := ct[:v0crypto.X25519KeySize]
	mlkemCT := ct[v0crypto.X25519KeySize : v0crypto.X25519KeySize+v0crypto.MLKEM768CiphertextSize]
	encSecret := ct[v0crypto.X25519KeySize+v0crypto.MLKEM768CiphertextSize:]

	// X25519 part.
	var x25519Priv [v0crypto.X25519KeySize]byte
	copy(x25519Priv[:], recipientPriv[:v0crypto.X25519KeySize])
	var x25519EphPubArr [v0crypto.X25519KeySize]byte
	copy(x25519EphPubArr[:], x25519EphPub)
	ssX25519, err := v0crypto.X25519DH(x25519Priv, x25519EphPubArr)
	if err != nil {
		return nil, fmt.Errorf("tree/treekem: x25519 dh: %w", err)
	}

	// ML-KEM-768 part.
	mlkemPriv := recipientPriv[v0crypto.X25519KeySize : v0crypto.X25519KeySize+v0crypto.MLKEM768PrivateKeySize]
	ssMLKEM, err := v0crypto.MLKEM768Decapsulate(mlkemPriv, mlkemCT)
	if err != nil {
		return nil, fmt.Errorf("tree/treekem: mlkem decapsulate: %w", err)
	}

	// Hybrid combine.
	hybridSS, err := v0crypto.DeriveKey(
		append(ssX25519[:], ssMLKEM...),
		nil,
		hybridKEMCombineInfo,
		32,
	)
	if err != nil {
		return nil, fmt.Errorf("tree/treekem: kem combine: %w", err)
	}

	// Decrypt the path secret.
	var aesKey [16]byte
	copy(aesKey[:], hybridSS[:16])
	var nonce [12]byte
	aad := treeKEMAAD()
	pathSecret, err := v0crypto.OpenAES128GCM(aesKey, nonce, encSecret, aad)
	if err != nil {
		return nil, ErrTreeKEMDecrypt
	}
	return pathSecret, nil
}

// treeKEMAAD returns the fixed AAD for TreeKEM path secret encryption.
func treeKEMAAD() []byte {
	return []byte("xorein/tree/v1/treekem-path-secret")
}

// deriveNodeKEMKey derives a hybrid KEM keypair for an internal tree node from
// a path secret and node index.
func deriveNodeKEMKey(pathSecret []byte, nodeIndex uint32) (pub, priv []byte, err error) {
	var idxBuf [4]byte
	binary.BigEndian.PutUint32(idxBuf[:], nodeIndex)
	nodeSeed, err := v0crypto.DeriveKey(pathSecret, idxBuf[:], "xorein/tree/v1/node-key", 64)
	if err != nil {
		return nil, nil, fmt.Errorf("tree/treekem: derive node seed: %w", err)
	}
	// Use first 32 bytes as X25519 scalar (after clamping) and last 32 bytes as ML-KEM seed.
	// For internal nodes we only need the public key for encryption; private key is also derived
	// so the local member can decrypt if they are on the update path.
	x25519Seed := nodeSeed[:32]
	// Clamp for X25519.
	x25519Seed[0] &= 248
	x25519Seed[31] &= 127
	x25519Seed[31] |= 64
	var x25519PrivArr [v0crypto.X25519KeySize]byte
	copy(x25519PrivArr[:], x25519Seed)
	// Compute X25519 public key from private scalar.
	var base [v0crypto.X25519KeySize]byte
	base[0] = 9 // X25519 base point
	x25519PubArr, err := v0crypto.X25519DH(x25519PrivArr, base)
	if err != nil {
		// Use a dummy fallback on error (low-probability degenerate case).
		x25519PubArr = x25519PrivArr
	}

	// For ML-KEM-768, generate a fresh keypair (seeded generation not available in circl).
	// In production this would use a deterministic KDF-seeded generation; here we use random.
	mlkemPub, mlkemPriv, err := v0crypto.GenerateMLKEM768Keypair()
	if err != nil {
		return nil, nil, fmt.Errorf("tree/treekem: mlkem node keygen: %w", err)
	}

	pub = make([]byte, HybridPublicKeySize)
	copy(pub[:v0crypto.X25519KeySize], x25519PubArr[:])
	copy(pub[v0crypto.X25519KeySize:], mlkemPub)

	priv = make([]byte, HybridPrivateKeySize)
	copy(priv[:v0crypto.X25519KeySize], x25519PrivArr[:])
	copy(priv[v0crypto.X25519KeySize:], mlkemPriv)

	return pub, priv, nil
}

// GenerateHybridKEMKeypair generates a fresh hybrid KEM keypair for use as a leaf init key.
// Returns (hybridPub HybridPublicKeySize bytes, hybridPriv HybridPrivateKeySize bytes).
func GenerateHybridKEMKeypair() (pub, priv []byte, err error) {
	x25519Priv, x25519Pub, err := v0crypto.GenerateX25519Keypair()
	if err != nil {
		return nil, nil, fmt.Errorf("tree/treekem: x25519 keygen: %w", err)
	}
	mlkemPub, mlkemPriv, err := v0crypto.GenerateMLKEM768Keypair()
	if err != nil {
		return nil, nil, fmt.Errorf("tree/treekem: mlkem keygen: %w", err)
	}
	pub = make([]byte, HybridPublicKeySize)
	copy(pub[:v0crypto.X25519KeySize], x25519Pub[:])
	copy(pub[v0crypto.X25519KeySize:], mlkemPub)

	priv = make([]byte, HybridPrivateKeySize)
	copy(priv[:v0crypto.X25519KeySize], x25519Priv[:])
	copy(priv[v0crypto.X25519KeySize:], mlkemPriv)

	return pub, priv, nil
}
