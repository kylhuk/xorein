package phase7

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrDuplicateMessage = errors.New("phase7: duplicate message")
	ErrInvalidSignature = errors.New("phase7: invalid signature")
)

type EncryptedMessage struct {
	Sender     ParticipantID
	Sequence   uint64
	Nonce      []byte
	Ciphertext []byte
	Signature  []byte
}

type Pipeline struct {
	mu          sync.Mutex
	seen        map[string]struct{}
	secret      []byte
	signer      ed25519.PrivateKey
	participant ParticipantID
}

func NewPipeline(id ParticipantID, state *KeyState) *Pipeline {
	return &Pipeline{
		seen:        make(map[string]struct{}),
		secret:      cloneKey(state.MLSSecret),
		signer:      state.Signer,
		participant: id,
	}
}

func (p *Pipeline) Send(seq uint64, plaintext []byte) (*EncryptedMessage, error) {
	nonce, err := randomNonce(12)
	if err != nil {
		return nil, fmt.Errorf("phase7: nonce generation failed: %w", err)
	}
	ciphertext, err := encrypt(p.secret, nonce, plaintext)
	if err != nil {
		return nil, err
	}
	sig := ed25519.Sign(p.signer, append(nonce, ciphertext...))
	return &EncryptedMessage{
		Sender:     p.participant,
		Sequence:   seq,
		Nonce:      nonce,
		Ciphertext: ciphertext,
		Signature:  sig,
	}, nil
}

func (p *Pipeline) Receive(msg *EncryptedMessage, secret []byte, verifier ed25519.PublicKey) ([]byte, error) {
	if !ed25519.Verify(verifier, append(msg.Nonce, msg.Ciphertext...), msg.Signature) {
		return nil, ErrInvalidSignature
	}
	id := replayID(msg.Sender, msg.Sequence)
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.seen[id]; ok {
		return nil, ErrDuplicateMessage
	}
	p.seen[id] = struct{}{}
	return decrypt(secret, msg.Nonce, msg.Ciphertext)
}

func encrypt(key, nonce, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Seal(nil, nonce, plaintext, nil), nil
}

func decrypt(key, nonce, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func randomNonce(size int) ([]byte, error) {
	nonce := make([]byte, size)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

func replayID(sender ParticipantID, seq uint64) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], seq)
	return string(sender) + ":" + string(buf[:])
}
