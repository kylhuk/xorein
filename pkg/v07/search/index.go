package search

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var ErrDuplicateDocument = errors.New("search.document.duplicate")

type Document struct {
	ID        string
	Scope     string
	Body      string
	From      string
	HasFile   bool
	HasLink   bool
	CreatedAt time.Time
}

type Index struct {
	documents  map[string]Document
	migrations []Migration
}

type Migration struct {
	ID        string
	AppliedAt time.Time
}

func NewIndex() *Index {
	return &Index{documents: make(map[string]Document)}
}

func (i *Index) Insert(doc Document) error {
	if doc.ID == "" {
		return fmt.Errorf("document id required")
	}
	if _, ok := i.documents[doc.ID]; ok {
		return ErrDuplicateDocument
	}
	i.documents[doc.ID] = doc
	return nil
}

func (i *Index) Update(doc Document) error {
	if doc.ID == "" {
		return fmt.Errorf("document id required")
	}
	i.documents[doc.ID] = doc
	return nil
}

func (i *Index) Remove(id string) {
	delete(i.documents, id)
}

func (i *Index) Query(scope string, rawQuery string, filters QueryFilters, pagination Pagination) ([]Document, error) {
	if _, err := NormalizeQuery(rawQuery, filters); err != nil {
		return nil, err
	}
	textQuery := strings.TrimSpace(rawQuery)
	page := NormalizePagination(pagination.Limit, pagination.Offset)
	results := make([]Document, 0, len(i.documents))
	for _, doc := range i.documents {
		if scope != "" && doc.Scope != scope {
			continue
		}
		if filters.FromUser != "" && doc.From != filters.FromUser {
			continue
		}
		if filters.HasFile && !doc.HasFile {
			continue
		}
		if filters.HasLink && !doc.HasLink {
			continue
		}
		if !filters.Range[0].IsZero() && !filters.Range[1].IsZero() {
			if doc.CreatedAt.Before(filters.Range[0]) || doc.CreatedAt.After(filters.Range[1]) {
				continue
			}
		}
		if !matchesText(doc, textQuery) {
			continue
		}
		results = append(results, doc)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})
	end := page.Offset + page.Limit
	if end > len(results) {
		end = len(results)
	}
	if page.Offset > len(results) {
		return []Document{}, nil
	}
	return results[page.Offset:end], nil
}

func (i *Index) ScopeInfo(scope string) int {
	count := 0
	for _, doc := range i.documents {
		if scope == "" || doc.Scope == scope {
			count++
		}
	}
	return count
}

func (i *Index) ApplyMigration(id string) Migration {
	mig := Migration{ID: id, AppliedAt: time.Now().UTC()}
	i.migrations = append(i.migrations, mig)
	return mig
}

func (i *Index) Migrations() []Migration {
	copies := append([]Migration(nil), i.migrations...)
	sort.Slice(copies, func(a, b int) bool {
		return copies[a].AppliedAt.Before(copies[b].AppliedAt)
	})
	return copies
}

func (i *Index) Rebuild() []Document {
	docs := make([]Document, 0, len(i.documents))
	for _, doc := range i.documents {
		docs = append(docs, doc)
	}
	sort.Slice(docs, func(a, b int) bool {
		return docs[a].CreatedAt.Before(docs[b].CreatedAt)
	})
	return docs
}

func (i *Index) MigrationStatus() string {
	if len(i.migrations) == 0 {
		return "pending"
	}
	return "applied"
}

func EnsureScope(scope string) error {
	if scope == "" {
		return errors.New("search.scope.required")
	}
	return nil
}

func matchesText(doc Document, query string) bool {
	if query == "" {
		return true
	}
	needle := strings.ToLower(query)
	body := strings.ToLower(doc.Body)
	return strings.Contains(body, needle)
}
