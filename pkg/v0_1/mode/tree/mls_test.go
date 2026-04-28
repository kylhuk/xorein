package tree_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
	tree "github.com/aether/code_aether/pkg/v0_1/mode/tree"
)

// TestCiphersuiteConstants verifies the ciphersuite ID and label.
func TestCiphersuiteConstants(t *testing.T) {
	if tree.CiphersuiteID != 0xFF01 {
		t.Fatalf("CiphersuiteID: want 0xFF01, got 0x%04X", tree.CiphersuiteID)
	}
	if tree.CiphersuiteLabel == "" {
		t.Fatal("CiphersuiteLabel must not be empty")
	}
}

// TestKeyPackageGenerateVerify verifies that a generated KeyPackage passes verification.
func TestKeyPackageGenerateVerify(t *testing.T) {
	edPub, edPriv, err := v0crypto.GenerateEd25519Keypair()
	if err != nil {
		t.Fatalf("generate ed25519: %v", err)
	}
	mldsaPub, mldsaPriv, err := v0crypto.GenerateMLDSA65Keypair()
	if err != nil {
		t.Fatalf("generate mldsa: %v", err)
	}

	kp, hybridPriv, err := tree.GenerateKeyPackage("alice@test", edPriv, mldsaPriv, edPub, mldsaPub)
	if err != nil {
		t.Fatalf("GenerateKeyPackage: %v", err)
	}
	if len(hybridPriv) != tree.HybridPrivateKeySize {
		t.Fatalf("hybridPriv size: want %d, got %d", tree.HybridPrivateKeySize, len(hybridPriv))
	}

	if kp.CiphersuiteID != tree.CiphersuiteID {
		t.Fatalf("want ciphersuite 0xFF01, got 0x%04X", kp.CiphersuiteID)
	}
	if len(kp.InitKey) != v0crypto.X25519KeySize {
		t.Fatalf("want init key size %d, got %d", v0crypto.X25519KeySize, len(kp.InitKey))
	}
	if len(kp.MLKEMPub) != v0crypto.MLKEM768PublicKeySize {
		t.Fatalf("want mlkem pub size %d, got %d", v0crypto.MLKEM768PublicKeySize, len(kp.MLKEMPub))
	}

	if err := tree.VerifyKeyPackage(kp); err != nil {
		t.Fatalf("VerifyKeyPackage: %v", err)
	}
}

// TestKeyPackageTamperedSignatureFails verifies tampered KeyPackages are rejected.
func TestKeyPackageTamperedSignatureFails(t *testing.T) {
	edPub, edPriv, _ := v0crypto.GenerateEd25519Keypair()
	mldsaPub, mldsaPriv, _ := v0crypto.GenerateMLDSA65Keypair()

	kp, _, err := tree.GenerateKeyPackage("alice@test", edPriv, mldsaPriv, edPub, mldsaPub)
	if err != nil {
		t.Fatalf("GenerateKeyPackage: %v", err)
	}

	// Tamper the signature.
	tampered := append([]byte(nil), kp.Signature...)
	tampered[0] ^= 0xFF
	kp.Signature = tampered

	if err := tree.VerifyKeyPackage(kp); err == nil {
		t.Fatal("expected verification failure for tampered signature")
	}
}

// TestEpochSecretsDerive verifies DeriveEpochSecrets produces distinct keys.
func TestEpochSecretsDerive(t *testing.T) {
	commitSecret := make([]byte, 32)
	rand.Read(commitSecret)

	es1, err := tree.DeriveEpochSecrets(commitSecret, 0, "group-1")
	if err != nil {
		t.Fatalf("DeriveEpochSecrets epoch 0: %v", err)
	}
	es2, err := tree.DeriveEpochSecrets(commitSecret, 1, "group-1")
	if err != nil {
		t.Fatalf("DeriveEpochSecrets epoch 1: %v", err)
	}

	// Different epochs must produce different secrets.
	if bytes.Equal(es1.HandshakeKey[:], es2.HandshakeKey[:]) {
		t.Fatal("HandshakeKey must differ across epochs")
	}
	if bytes.Equal(es1.ApplicationKey[:], es2.ApplicationKey[:]) {
		t.Fatal("ApplicationKey must differ across epochs")
	}
	if bytes.Equal(es1.ExporterSecret[:], es2.ExporterSecret[:]) {
		t.Fatal("ExporterSecret must differ across epochs")
	}
}

// TestMLSExporterDeterministic verifies MLSExporter is deterministic.
func TestMLSExporterDeterministic(t *testing.T) {
	commitSecret := make([]byte, 32)
	rand.Read(commitSecret)

	es, err := tree.DeriveEpochSecrets(commitSecret, 0, "grp")
	if err != nil {
		t.Fatalf("DeriveEpochSecrets: %v", err)
	}

	out1, err := es.MLSExporter("test-label", []byte("context"), 32)
	if err != nil {
		t.Fatalf("MLSExporter 1: %v", err)
	}
	out2, err := es.MLSExporter("test-label", []byte("context"), 32)
	if err != nil {
		t.Fatalf("MLSExporter 2: %v", err)
	}

	if !bytes.Equal(out1, out2) {
		t.Fatal("MLSExporter must be deterministic")
	}

	// Different label → different output.
	out3, err := es.MLSExporter("other-label", []byte("context"), 32)
	if err != nil {
		t.Fatalf("MLSExporter 3: %v", err)
	}
	if bytes.Equal(out1, out3) {
		t.Fatal("MLSExporter: different label must produce different output")
	}
}

// TestMediaShieldExporterLabel verifies the MediaShield exporter uses the correct label.
func TestMediaShieldExporterLabel(t *testing.T) {
	commitSecret := make([]byte, 32)
	rand.Read(commitSecret)

	es, _ := tree.DeriveEpochSecrets(commitSecret, 0, "grp")
	msk, err := es.DeriveMediaShieldKey()
	if err != nil {
		t.Fatalf("DeriveMediaShieldKey: %v", err)
	}
	if len(msk) != 32 {
		t.Fatalf("DeriveMediaShieldKey: want 32 bytes, got %d", len(msk))
	}
	// Verify it matches manual MLSExporter call with the spec label.
	msk2, err := es.MLSExporter(tree.LabelMediaShieldExporter, []byte{}, 32)
	if err != nil {
		t.Fatalf("MLSExporter manual: %v", err)
	}
	if !bytes.Equal(msk, msk2) {
		t.Fatal("DeriveMediaShieldKey must match MLSExporter with label")
	}
}

// TestTreeKEMRoundtrip verifies UpdatePath / ProcessUpdatePath for a 2-member group.
func TestTreeKEMRoundtrip(t *testing.T) {
	// Alice and Bob each generate hybrid KEM keypairs.
	alicePub, alicePriv, err := tree.GenerateHybridKEMKeypair()
	if err != nil {
		t.Fatalf("alice keygen: %v", err)
	}
	bobPub, bobPriv, err := tree.GenerateHybridKEMKeypair()
	if err != nil {
		t.Fatalf("bob keygen: %v", err)
	}

	ratchetTree := tree.NewRatchetTree("test-group", [][]byte{alicePub, bobPub})

	// Alice performs an UpdatePath.
	_, aliceNewPriv, err := tree.GenerateHybridKEMKeypair()
	if err != nil {
		t.Fatalf("alice new keygen: %v", err)
	}
	_ = aliceNewPriv

	// Alice updates path with her existing keypair.
	pathSecret, encryptedPathSecrets, err := ratchetTree.UpdatePath(0, alicePriv, alicePub)
	if err != nil {
		t.Fatalf("UpdatePath: %v", err)
	}
	if len(pathSecret) != 32 {
		t.Fatalf("pathSecret: want 32 bytes, got %d", len(pathSecret))
	}
	if len(encryptedPathSecrets) == 0 {
		t.Fatal("want encrypted path secrets")
	}

	// Bob processes the UpdatePath using his private key.
	// For a 2-member group (1 co-path node), the first non-nil encrypted path secret
	// is targeted at Bob.
	var bobEncSecret []byte
	for _, eps := range encryptedPathSecrets {
		if eps != nil {
			bobEncSecret = eps
			break
		}
	}
	if bobEncSecret == nil {
		t.Fatal("no encrypted path secret for bob")
	}

	commitSecret, err := ratchetTree.ProcessUpdatePath(1, bobEncSecret, bobPriv)
	if err != nil {
		t.Fatalf("ProcessUpdatePath: %v", err)
	}
	if !bytes.Equal(commitSecret, pathSecret) {
		t.Fatalf("commit secret mismatch:\n  want %x\n  got  %x", pathSecret, commitSecret)
	}
}

// TestWelcomeApplyRoundtrip verifies NewWelcome / ApplyWelcome can exchange epoch secrets.
func TestWelcomeApplyRoundtrip(t *testing.T) {
	// Generate Bob's KeyPackage (new member).
	edPub, edPriv, _ := v0crypto.GenerateEd25519Keypair()
	mldsaPub, mldsaPriv, _ := v0crypto.GenerateMLDSA65Keypair()
	bobKP, bobHybridPriv, err := tree.GenerateKeyPackage("bob@test", edPriv, mldsaPriv, edPub, mldsaPub)
	if err != nil {
		t.Fatalf("generate bob kp: %v", err)
	}

	// Alice's group.
	aliceGroup, _ := tree.NewGroup("test-group", tree.Member{PeerID: "alice@test"})

	// Generate epoch secrets (simulating an MLS commit).
	commitSecret := make([]byte, 32)
	rand.Read(commitSecret)
	es, err := tree.DeriveEpochSecrets(commitSecret, aliceGroup.CurrentEpoch.EpochID, "test-group")
	if err != nil {
		t.Fatalf("DeriveEpochSecrets: %v", err)
	}

	// Build Welcome for Bob.
	welcome, err := tree.NewWelcome(aliceGroup, es, bobKP)
	if err != nil {
		t.Fatalf("NewWelcome: %v", err)
	}

	// Bob applies the Welcome.
	bobGroup, err := tree.ApplyWelcome(welcome, bobKP, bobHybridPriv)
	if err != nil {
		t.Fatalf("ApplyWelcome: %v", err)
	}

	// The epoch secrets should match.
	if bobGroup.MLSEpoch == nil {
		t.Fatal("bob MLSEpoch must not be nil after ApplyWelcome")
	}
	if !bytes.Equal(bobGroup.MLSEpoch.HandshakeKey[:], es.HandshakeKey[:]) {
		t.Fatal("HandshakeKey mismatch after Welcome/Apply")
	}
	if !bytes.Equal(bobGroup.MLSEpoch.ApplicationKey[:], es.ApplicationKey[:]) {
		t.Fatal("ApplicationKey mismatch after Welcome/Apply")
	}
	if !bytes.Equal(bobGroup.MLSEpoch.ExporterSecret[:], es.ExporterSecret[:]) {
		t.Fatal("ExporterSecret mismatch after Welcome/Apply")
	}
}

// TestMLSMessageSignVerify tests hybrid signing and verification of MLSMessage.
func TestMLSMessageSignVerify(t *testing.T) {
	edPub, edPriv, err := v0crypto.GenerateEd25519Keypair()
	if err != nil {
		t.Fatalf("generate ed25519: %v", err)
	}
	mldsaPub, mldsaPriv, err := v0crypto.GenerateMLDSA65Keypair()
	if err != nil {
		t.Fatalf("generate mldsa: %v", err)
	}

	msg := &tree.MLSMessage{
		Version:     tree.MLSVersion,
		ContentType: tree.ContentTypeCommit,
		GroupID:     "group-xyz",
		EpochID:     42,
		Payload:     []byte("commit payload"),
	}

	if err := tree.SignMLSMessage(msg, ed25519.PrivateKey(edPriv), mldsaPriv); err != nil {
		t.Fatalf("SignMLSMessage: %v", err)
	}
	if len(msg.Signature) != v0crypto.HybridSignatureSize {
		t.Fatalf("signature size: want %d, got %d", v0crypto.HybridSignatureSize, len(msg.Signature))
	}
	if err := tree.VerifyMLSMessage(msg, ed25519.PublicKey(edPub), mldsaPub); err != nil {
		t.Fatalf("VerifyMLSMessage: %v", err)
	}
}
