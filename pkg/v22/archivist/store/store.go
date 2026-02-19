package store

import (
	"time"
)

type SpaceID string
type ChannelID string
type SegmentID string

type SegmentKey struct {
	Space   SpaceID
	Channel ChannelID
	Segment SegmentID
}

type Segment struct {
	SizeBytes  int64
	InsertedAt time.Time
}

type StoreReason string

const (
	ReasonQuotaExceeded   StoreReason = "QUOTA_EXCEEDED"
	ReasonRetentionPolicy StoreReason = "RETENTION_POLICY"
	ReasonSegmentTooLarge StoreReason = "SEGMENT_TOO_LARGE"
)

type StoreError struct {
	Reason StoreReason
}

func (s StoreError) Error() string {
	return string(s.Reason)
}

type Config struct {
	QuotaPerSpace  map[SpaceID]int64
	ChannelCap     map[ChannelID]int64
	MaxSegmentSize int64
	Retention      time.Duration
	Now            func() time.Time
}

type Store struct {
	config       Config
	data         map[SegmentKey]Segment
	spaceUsage   map[SpaceID]int64
	channelUsage map[ChannelID]int64
}

func NewStore(cfg Config) *Store {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.QuotaPerSpace == nil {
		cfg.QuotaPerSpace = make(map[SpaceID]int64)
	}
	if cfg.ChannelCap == nil {
		cfg.ChannelCap = make(map[ChannelID]int64)
	}

	return &Store{
		config:       cfg,
		data:         make(map[SegmentKey]Segment),
		spaceUsage:   make(map[SpaceID]int64),
		channelUsage: make(map[ChannelID]int64),
	}
}

func (s *Store) Put(space SpaceID, channel ChannelID, segment SegmentID, sizeBytes int64) error {
	if sizeBytes <= 0 {
		return nil
	}
	if s.config.MaxSegmentSize > 0 && sizeBytes > s.config.MaxSegmentSize {
		return StoreError{Reason: ReasonSegmentTooLarge}
	}
	if capBytes, ok := s.config.ChannelCap[channel]; ok && capBytes > 0 {
		if s.channelUsage[channel]+sizeBytes > capBytes {
			return StoreError{Reason: ReasonQuotaExceeded}
		}
	}
	if quota, ok := s.config.QuotaPerSpace[space]; ok && quota > 0 {
		if s.spaceUsage[space]+sizeBytes > quota {
			return StoreError{Reason: ReasonQuotaExceeded}
		}
	}

	key := SegmentKey{Space: space, Channel: channel, Segment: segment}
	now := s.config.Now()

	if prev, exists := s.data[key]; exists {
		s.spaceUsage[space] -= prev.SizeBytes
		s.channelUsage[channel] -= prev.SizeBytes
	}

	s.data[key] = Segment{SizeBytes: sizeBytes, InsertedAt: now}
	s.spaceUsage[space] += sizeBytes
	s.channelUsage[channel] += sizeBytes
	return nil
}

type PrunedSegment struct {
	Key    SegmentKey
	Reason StoreReason
	Size   int64
}

func (s *Store) Prune() []PrunedSegment {
	if s.config.Retention <= 0 {
		return nil
	}
	now := s.config.Now()
	cutoff := now.Add(-s.config.Retention)
	var pruned []PrunedSegment

	for key, segment := range s.data {
		if segment.InsertedAt.Before(cutoff) {
			pruned = append(pruned, PrunedSegment{Key: key, Reason: ReasonRetentionPolicy, Size: segment.SizeBytes})
			s.spaceUsage[key.Space] -= segment.SizeBytes
			s.channelUsage[key.Channel] -= segment.SizeBytes
			delete(s.data, key)
		}
	}

	if len(pruned) > 0 {
		return pruned
	}
	return nil
}

func (s *Store) SpaceUsage(space SpaceID) int64 {
	return s.spaceUsage[space]
}

func (s *Store) ChannelUsage(channel ChannelID) int64 {
	return s.channelUsage[channel]
}

func (s *Store) SegmentCount() int {
	return len(s.data)
}
