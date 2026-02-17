package search

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidDateRange = errors.New("search.invalid_date_range")

type QueryFilters struct {
	FromUser string
	Range    [2]time.Time
	HasFile  bool
	HasLink  bool
}

func ValidateFilters(filters QueryFilters) error {
	startZero := filters.Range[0].IsZero()
	endZero := filters.Range[1].IsZero()
	if startZero != endZero {
		return ErrInvalidDateRange
	}
	if !startZero && filters.Range[1].Before(filters.Range[0]) {
		return ErrInvalidDateRange
	}
	return nil
}

func NormalizeQuery(raw string, filters QueryFilters) (string, error) {
	if err := ValidateFilters(filters); err != nil {
		return "", err
	}

	parts := []string{strings.TrimSpace(raw)}
	if filters.FromUser != "" {
		parts = append(parts, fmt.Sprintf("from:%s", filters.FromUser))
	}
	if !filters.Range[0].IsZero() && !filters.Range[1].IsZero() {
		parts = append(parts, fmt.Sprintf("after:%s before:%s", filters.Range[0].Format(time.RFC3339), filters.Range[1].Format(time.RFC3339)))
	}
	if filters.HasFile {
		parts = append(parts, "has:file")
	}
	if filters.HasLink {
		parts = append(parts, "has:link")
	}
	return strings.Join(parts, " "), nil
}

type Pagination struct {
	Limit  int
	Offset int
}

func NormalizePagination(limit, offset int) Pagination {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return Pagination{Limit: limit, Offset: offset}
}

func ScopeAuthorized(scopeID, actor string) bool {
	return strings.HasPrefix(actor, "user:") && strings.Contains(scopeID, "S7")
}
