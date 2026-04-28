package protocol

// Package protocol contains protocol family/version registry, feature-flag
// negotiation, and compatibility helpers for v0.1 and additive v0.2 contracts.
//
// The negotiation path uses canonical protocol IDs ordered from highest to
// lowest version, then applies a compatibility policy and optional deprecation
// guard to choose the best common protocol without breaking older peers.
