package merkle

import (
	"testing"
	"testing/quick"
)

func TestBuilderDeterministicRoot(t *testing.T) {
	seed := []string{"alpha", "beta", "gamma"}
	root1 := buildRoot(seed)
	root2 := buildRoot(seed)
	if root1 != root2 {
		t.Fatalf("expected deterministic root, got %s vs %s", root1, root2)
	}
}

func TestVerifyProofCorruptedLeaf(t *testing.T) {
	b := NewBuilder()
	b.AddLeaf([]byte("one"))
	b.AddLeaf([]byte("two"))
	root := b.Root()
	proof, err := b.Proof(0)
	if err != nil {
		t.Fatalf("proof generation failed: %v", err)
	}
	cases := []struct {
		name  string
		data  []byte
		proof []string
		total int
		want  error
	}{
		{name: "valid proof", data: []byte("one"), proof: proof, total: 2, want: nil},
		{name: "corrupted leaf", data: []byte("tampered"), proof: proof, total: 2, want: ErrDivergentRoot},
		{name: "invalid proof segment", data: []byte("one"), proof: []string{"zz"}, total: 2, want: ErrInvalidProof},
		{name: "invalid total", data: []byte("one"), proof: proof, total: 0, want: ErrInvalidProof},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := VerifyProof(root, tc.data, tc.proof, 0, tc.total)
			if tc.want == nil && err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			if tc.want != nil && err != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestBuildRootDeterministicProperty(t *testing.T) {
	values := []string{"alpha", "beta", "gamma"}
	b1 := NewBuilder()
	for _, v := range values {
		b1.AddLeaf([]byte(v))
	}
	root1 := b1.Root()
	b2 := NewBuilder()
	for _, v := range values {
		b2.AddLeaf([]byte(v))
	}
	root2 := b2.Root()
	if root1 != root2 {
		t.Fatalf("expected deterministic root, got %s vs %s", root1, root2)
	}
	// confirm rearranging leaf addition changes root predictably
	different := NewBuilder()
	for i := len(values) - 1; i >= 0; i-- {
		different.AddLeaf([]byte(values[i]))
	}
	if root1 == different.Root() {
		t.Fatalf("expected order-sensitive results")
	}
}

func TestBuildRootQuickProperty(t *testing.T) {
	prop := func(values []string) bool {
		if len(values) == 0 {
			values = []string{"seed"}
		}

		left := NewBuilder()
		right := NewBuilder()
		for _, value := range values {
			left.AddLeaf([]byte(value))
			right.AddLeaf([]byte(value))
		}

		return left.Root() == right.Root()
	}

	if err := quick.Check(prop, nil); err != nil {
		t.Fatalf("quick determinism property failed: %v", err)
	}
}

func buildRoot(values []string) string {
	b := NewBuilder()
	for _, v := range values {
		b.AddLeaf([]byte(v))
	}
	return b.Root()
}
