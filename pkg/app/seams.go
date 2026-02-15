package app

import "context"

// Seams captures the minimal interfaces the Phase 2 foundation exposes between
// the higher-level application wiring and each responsibility center. Each
// interface stays intentionally small to keep the scaffold testable while specific
// implementations live under pkg/{protocol,network,crypto,storage,ui}.
type Seams struct {
	Protocol ProtocolSeam
	Network  NetworkSeam
	Crypto   CryptoSeam
	Storage  StorageSeam
	UI       UISeam
}

// ProtocolSeam defines the slice of behaviour app-level code relies on from the
// protocol package. The concrete flow will be determined in later Phase 2
// deliverables.
type ProtocolSeam interface {
	AlignProtocolLifecycle(ctx context.Context) error
}

// NetworkSeam represents the guarantees expected from the network package.
type NetworkSeam interface {
	EnsureTransportReady(ctx context.Context) error
}

// CryptoSeam abstracts any cryptographic helper the application wires in for P2.
type CryptoSeam interface {
	ProtectPayload(ctx context.Context, input []byte) ([]byte, error)
}

// StorageSeam covers the persistence operations needed during Phase 2 scaffolding.
type StorageSeam interface {
	PersistBlob(ctx context.Context, key string, data []byte) error
	LoadBlob(ctx context.Context, key string) ([]byte, error)
}

// UISeam exposes the UI-adjacent entry points the app will invoke in Phase 2.
type UISeam interface {
	RenderPlaceholder(ctx context.Context) error
}
