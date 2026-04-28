package seal

import (
	"encoding/binary"
	"errors"
	"fmt"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
)

const (
	// HeaderSize is the fixed binary message header size in bytes per spec 11 §3.3.
	HeaderSize = 53 // 1 + 4 + 4 + 32 + 12

	// MaxSkippedMessages is the maximum number of out-of-order messages retained per spec 11 §3.4.
	MaxSkippedMessages = 1000

	headerVersion = 0x01
)

// ErrMaxSkippedExceeded is returned when a message counter would exceed the skip limit.
var ErrMaxSkippedExceeded = errors.New("seal: skipped message count exceeds maximum 1000")

// ErrBadHeaderVersion is returned for unknown header versions.
var ErrBadHeaderVersion = errors.New("seal: unsupported message header version")

// SkipKey identifies a skipped message by its ratchet pub key and counter.
type SkipKey struct {
	RatchetPub [32]byte
	Counter    uint32
}

// RatchetState is the Double Ratchet state for a Seal session.
type RatchetState struct {
	RootKey          [32]byte
	SendChainKey     [32]byte
	RecvChainKey     [32]byte
	SendCounter      uint32
	RecvCounter      uint32
	PrevSendChainLen uint32
	SendRatchetPriv  [32]byte
	SendRatchetPub   [32]byte
	RemoteRatchetPub [32]byte
	SkipList         map[SkipKey][32]byte
}

// Encrypt encrypts plaintext, returns a 53-byte header and ciphertext (with AEAD tag).
// The returned header is the AEAD associated data and must be sent alongside the ciphertext.
func Encrypt(s *RatchetState, plaintext []byte) (header [HeaderSize]byte, ciphertext []byte, err error) {
	// Derive message key.
	mk, nextChain, err := advanceSendChain(s)
	if err != nil {
		return header, nil, fmt.Errorf("ratchet encrypt: %w", err)
	}

	// Build header.
	nonce, err := v0crypto.RandomNonce12()
	if err != nil {
		return header, nil, fmt.Errorf("ratchet encrypt: nonce: %w", err)
	}
	header = buildHeader(s.SendCounter, s.PrevSendChainLen, s.SendRatchetPub, nonce)

	// Encrypt.
	var k [32]byte
	copy(k[:], mk[:32])
	ct, err := v0crypto.SealChaCha20Poly1305(k, nonce, plaintext, header[:])
	if err != nil {
		return header, nil, fmt.Errorf("ratchet encrypt: seal: %w", err)
	}

	// Advance state.
	s.SendCounter++
	s.SendChainKey = nextChain
	return header, ct, nil
}

// Decrypt decrypts a message given its 53-byte header and ciphertext.
func Decrypt(s *RatchetState, header [HeaderSize]byte, ciphertext []byte) ([]byte, error) {
	if header[0] != headerVersion {
		return nil, ErrBadHeaderVersion
	}
	counter := binary.BigEndian.Uint32(header[1:5])
	var ratchetPub [32]byte
	copy(ratchetPub[:], header[9:41])
	var nonce [12]byte
	copy(nonce[:], header[41:53])

	// Check skip list first.
	sk := SkipKey{RatchetPub: ratchetPub, Counter: counter}
	if mk, ok := s.SkipList[sk]; ok {
		delete(s.SkipList, sk)
		return openMessage(mk, nonce, header, ciphertext)
	}

	// Detect DH ratchet step.
	needsDHStep := ratchetPub != s.RemoteRatchetPub

	if needsDHStep {
		// Save skipped messages from current recv chain.
		prevChainLen := binary.BigEndian.Uint32(header[5:9])
		if err := skipMessages(s, s.RemoteRatchetPub, prevChainLen); err != nil {
			return nil, err
		}
		// Perform DH ratchet step.
		if err := dhRatchetStep(s, ratchetPub); err != nil {
			return nil, fmt.Errorf("ratchet decrypt: DH step: %w", err)
		}
	}

	// Skip messages in current recv chain up to counter.
	if err := skipMessages(s, ratchetPub, counter); err != nil {
		return nil, err
	}

	// Derive message key from current recv chain.
	mk, nextChain, err := advanceRecvChain(s)
	if err != nil {
		return nil, fmt.Errorf("ratchet decrypt: advance chain: %w", err)
	}
	s.RecvChainKey = nextChain
	s.RecvCounter++

	return openMessage(mk, nonce, header, ciphertext)
}

// advanceSendChain derives the next message key and advances the send chain.
func advanceSendChain(s *RatchetState) (mk [32]byte, nextChain [32]byte, err error) {
	okm, err := v0crypto.DeriveKey(s.SendChainKey[:], []byte{0x01}, v0crypto.LabelSealMessageKey, 64)
	if err != nil {
		return mk, nextChain, err
	}
	copy(mk[:], okm[:32])
	copy(nextChain[:], okm[32:])
	return mk, nextChain, nil
}

// advanceRecvChain derives the next message key from the recv chain.
func advanceRecvChain(s *RatchetState) (mk [32]byte, nextChain [32]byte, err error) {
	okm, err := v0crypto.DeriveKey(s.RecvChainKey[:], []byte{0x01}, v0crypto.LabelSealMessageKey, 64)
	if err != nil {
		return mk, nextChain, err
	}
	copy(mk[:], okm[:32])
	copy(nextChain[:], okm[32:])
	return mk, nextChain, nil
}

// skipMessages advances the recv chain to targetCounter, storing skipped message keys.
func skipMessages(s *RatchetState, ratchetPub [32]byte, targetCounter uint32) error {
	if len(s.SkipList)+int(targetCounter-s.RecvCounter) > MaxSkippedMessages {
		return ErrMaxSkippedExceeded
	}
	for s.RecvCounter < targetCounter {
		mk, nextChain, err := advanceRecvChain(s)
		if err != nil {
			return err
		}
		s.SkipList[SkipKey{RatchetPub: ratchetPub, Counter: s.RecvCounter}] = mk
		s.RecvChainKey = nextChain
		s.RecvCounter++
	}
	return nil
}

// dhRatchetStep performs a DH ratchet step when a new ratchet key is seen.
// Per spec 11 §3.5.
func dhRatchetStep(s *RatchetState, newRemotePub [32]byte) error {
	// Derive new recv chain.
	dhOut, err := v0crypto.X25519DH(s.SendRatchetPriv, newRemotePub)
	if err != nil {
		return fmt.Errorf("DH for recv chain: %w", err)
	}
	okm1, err := v0crypto.DeriveKey(s.RootKey[:], dhOut[:], v0crypto.LabelSealRatchetStep, 64)
	if err != nil {
		return err
	}
	var newRoot [32]byte
	copy(newRoot[:], okm1[:32])
	copy(s.RecvChainKey[:], okm1[32:])
	s.RecvCounter = 0

	// Generate new sending ratchet key.
	newSendPriv, newSendPub, err := v0crypto.GenerateX25519Keypair()
	if err != nil {
		return fmt.Errorf("generate new send ratchet: %w", err)
	}

	// Derive new send chain.
	dhOut2, err := v0crypto.X25519DH(newSendPriv, newRemotePub)
	if err != nil {
		return fmt.Errorf("DH for send chain: %w", err)
	}
	okm2, err := v0crypto.DeriveKey(newRoot[:], dhOut2[:], v0crypto.LabelSealRatchetStep, 64)
	if err != nil {
		return err
	}

	// Update state.
	s.PrevSendChainLen = s.SendCounter
	s.SendCounter = 0
	s.RootKey = newRoot
	copy(s.SendChainKey[:], okm2[32:])
	s.SendRatchetPriv = newSendPriv
	s.SendRatchetPub = newSendPub
	s.RemoteRatchetPub = newRemotePub
	return nil
}

// buildHeader constructs the 53-byte message header per spec 11 §3.3.
func buildHeader(counter, prevChainLen uint32, ratchetPub [32]byte, nonce [12]byte) [HeaderSize]byte {
	var h [HeaderSize]byte
	h[0] = headerVersion
	binary.BigEndian.PutUint32(h[1:5], counter)
	binary.BigEndian.PutUint32(h[5:9], prevChainLen)
	copy(h[9:41], ratchetPub[:])
	copy(h[41:53], nonce[:])
	return h
}

// openMessage decrypts ciphertext using ChaCha20-Poly1305.
func openMessage(mk [32]byte, nonce [12]byte, header [HeaderSize]byte, ct []byte) ([]byte, error) {
	pt, err := v0crypto.OpenChaCha20Poly1305(mk, nonce, ct, header[:])
	if err != nil {
		return nil, fmt.Errorf("ratchet decrypt: open: %w", err)
	}
	return pt, nil
}
