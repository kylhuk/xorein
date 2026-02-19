package replication

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Provider represents a storage provider in the r_blob catalog.
type Provider struct {
	ID      string
	Region  string
	ASN     string
	Healthy bool
}

// Config controls how replication targets are selected.
type Config struct {
	TargetReplicas  int
	PreferredRegion string
	AvoidSingleASN  bool
}

// RefusalCode is a deterministic refusal reason emitted by replication workflows.
type RefusalCode string

const (
	RefusalCodeInsufficientProviders RefusalCode = "insufficient_providers"
	RefusalCodePublishFailure        RefusalCode = "publish_failure"
	RefusalCodeRepairFailure         RefusalCode = "repair_failure"
)

// RefusalError encodes a deterministic refusal code and human reason.
type RefusalError struct {
	Code   RefusalCode
	Reason string
}

func (e RefusalError) Error() string {
	if e.Code == "" && e.Reason == "" {
		return "replication refusal"
	}
	if e.Reason == "" {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Reason)
}

func newRefusalError(code RefusalCode, reason string) error {
	return RefusalError{Code: code, Reason: reason}
}

// ReplicaProviderRecord tracks where a blob replica resides.
type ReplicaProviderRecord struct {
	ProviderID string
	Region     string
	ASN        string
	AssignedAt time.Time
}

// ReplicaMetadata tracks the recorded replica set for a blob.
type ReplicaMetadata struct {
	BlobHash       string
	Providers      []ReplicaProviderRecord
	PublishedAt    time.Time
	LastVerifiedAt time.Time
}

// Manager materializes deterministic publish/repair workflows.
type Manager struct {
	providers map[string]Provider
	config    Config
	metadata  map[string]ReplicaMetadata
	now       func() time.Time
}

// NewManager returns a replication manager with the provided catalog.
func NewManager(catalog []Provider, cfg Config) (*Manager, error) {
	if cfg.TargetReplicas <= 0 {
		return nil, fmt.Errorf("replication target must be positive")
	}
	providers := make(map[string]Provider, len(catalog))
	for _, p := range catalog {
		providers[p.ID] = p
	}
	return &Manager{
		providers: providers,
		config:    cfg,
		metadata:  make(map[string]ReplicaMetadata),
		now:       time.Now,
	}, nil
}

// SetProvider adds or updates a provider in the catalog.
func (m *Manager) SetProvider(provider Provider) {
	existing, ok := m.providers[provider.ID]
	if !ok {
		m.providers[provider.ID] = provider
		return
	}
	if provider.Region != "" {
		existing.Region = provider.Region
	}
	if provider.ASN != "" {
		existing.ASN = provider.ASN
	}
	existing.Healthy = provider.Healthy
	m.providers[provider.ID] = existing
}

// SetTimeSource replaces the time source (tests only).
func (m *Manager) SetTimeSource(now func() time.Time) {
	if now == nil {
		return
	}
	m.now = now
}

// Publish records a replica set for the given blob hash.
func (m *Manager) Publish(blobHash string) error {
	if blobHash == "" {
		return newRefusalError(RefusalCodePublishFailure, "blob hash required")
	}
	selected, err := m.pickProviders(m.config.TargetReplicas, nil, nil)
	if err != nil {
		return newRefusalError(RefusalCodeInsufficientProviders, err.Error())
	}
	now := m.now()
	records := make([]ReplicaProviderRecord, len(selected))
	for i, provider := range selected {
		records[i] = m.newRecord(provider, now)
	}
	m.metadata[blobHash] = ReplicaMetadata{
		BlobHash:       blobHash,
		Providers:      records,
		PublishedAt:    now,
		LastVerifiedAt: now,
	}
	return nil
}

// ReplicaMetadata returns metadata for the blob.
func (m *Manager) ReplicaMetadata(blobHash string) (ReplicaMetadata, bool) {
	meta, ok := m.metadata[blobHash]
	return meta, ok
}

// VerifyAndRepair ensures target replicas remain healthy.
func (m *Manager) VerifyAndRepair(blobHash string) error {
	meta, ok := m.metadata[blobHash]
	if !ok {
		return newRefusalError(RefusalCodeRepairFailure, "replica metadata missing")
	}
	recordByID := make(map[string]ReplicaProviderRecord, len(meta.Providers))
	healthyProviders := make([]Provider, 0, len(meta.Providers))
	usedASNs := make(map[string]struct{})
	for _, record := range meta.Providers {
		recordByID[record.ProviderID] = record
		provider, exists := m.providers[record.ProviderID]
		if !exists || !provider.Healthy {
			continue
		}
		healthyProviders = append(healthyProviders, provider)
		if provider.ASN != "" {
			usedASNs[provider.ASN] = struct{}{}
		}
	}
	now := m.now()
	records := make([]ReplicaProviderRecord, 0, m.config.TargetReplicas)
	for _, provider := range healthyProviders {
		if rec, ok := recordByID[provider.ID]; ok {
			records = append(records, rec)
			continue
		}
		records = append(records, m.newRecord(provider, now))
	}
	if len(records) > m.config.TargetReplicas {
		records = records[:m.config.TargetReplicas]
	}
	if len(records) < m.config.TargetReplicas {
		excludeIDs := make(map[string]struct{}, len(records))
		for _, r := range records {
			excludeIDs[r.ProviderID] = struct{}{}
		}
		needed := m.config.TargetReplicas - len(records)
		replacements, err := m.pickProviders(needed, excludeIDs, usedASNs)
		if err != nil {
			return newRefusalError(RefusalCodeRepairFailure, err.Error())
		}
		for _, provider := range replacements {
			records = append(records, m.newRecord(provider, now))
		}
	}
	if len(records) != m.config.TargetReplicas {
		return newRefusalError(RefusalCodeRepairFailure, "unable to reach target replica count")
	}
	meta.Providers = records
	meta.LastVerifiedAt = now
	m.metadata[blobHash] = meta
	return nil
}

func (m *Manager) newRecord(provider Provider, now time.Time) ReplicaProviderRecord {
	return ReplicaProviderRecord{
		ProviderID: provider.ID,
		Region:     provider.Region,
		ASN:        provider.ASN,
		AssignedAt: now,
	}
}

func (m *Manager) pickProviders(needed int, excludeIDs, existingASNs map[string]struct{}) ([]Provider, error) {
	candidates := make([]Provider, 0, len(m.providers))
	for _, provider := range m.providers {
		if !provider.Healthy {
			continue
		}
		if excludeIDs != nil {
			if _, ok := excludeIDs[provider.ID]; ok {
				continue
			}
		}
		candidates = append(candidates, provider)
	}
	sort.Slice(candidates, func(i, j int) bool {
		pi := candidates[i]
		pj := candidates[j]
		piPref := regionPriority(pi.Region, m.config.PreferredRegion)
		pjPref := regionPriority(pj.Region, m.config.PreferredRegion)
		if piPref != pjPref {
			return piPref < pjPref
		}
		if pi.Region != pj.Region {
			return pi.Region < pj.Region
		}
		return pi.ID < pj.ID
	})
	selected := make([]Provider, 0, needed)
	seenASNs := make(map[string]struct{}, len(existingASNs))
	for asn := range existingASNs {
		if asn != "" {
			seenASNs[asn] = struct{}{}
		}
	}
	for _, provider := range candidates {
		if len(selected) == needed {
			break
		}
		if m.config.AvoidSingleASN && provider.ASN != "" {
			if _, ok := seenASNs[provider.ASN]; ok {
				continue
			}
			seenASNs[provider.ASN] = struct{}{}
		}
		selected = append(selected, provider)
	}
	if len(selected) < needed {
		for _, provider := range candidates {
			if len(selected) == needed {
				break
			}
			if containsProvider(selected, provider) {
				continue
			}
			selected = append(selected, provider)
		}
	}
	if len(selected) < needed {
		return nil, fmt.Errorf("need %d providers but only %d available", needed, len(selected))
	}
	return selected, nil
}

func regionPriority(region, preferred string) int {
	if preferred != "" {
		if strings.EqualFold(region, preferred) {
			return 0
		}
		if region == "" {
			return 2
		}
	}
	if region != "" {
		return 1
	}
	return 2
}

func containsProvider(set []Provider, provider Provider) bool {
	for _, candidate := range set {
		if candidate.ID == provider.ID {
			return true
		}
	}
	return false
}
