package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Storage manages audit records on the local filesystem.
// Layout: baseDir/{audit-id}/meta.json + report.{html,json,md}
type Storage struct {
	baseDir string
	mu      sync.RWMutex
}

// NewStorage creates (or opens) the storage root directory.
func NewStorage(baseDir string) (*Storage, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir %q: %w", baseDir, err)
	}
	return &Storage{baseDir: baseDir}, nil
}

// AuditDir returns the directory path for a given audit ID.
func (s *Storage) AuditDir(id string) string {
	return filepath.Join(s.baseDir, id)
}

// ReportPath returns the full path to a report file of the given format.
func (s *Storage) ReportPath(id, format string) string {
	return filepath.Join(s.AuditDir(id), "report."+format)
}

// BaseDir exposes the configured root for external use.
func (s *Storage) BaseDir() string { return s.baseDir }

// Save writes an AuditRecord to disk (atomic-friendly: write then rename would be
// ideal, but MarshalIndent + WriteFile is sufficient for local tooling).
func (s *Storage) Save(record *AuditRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := s.AuditDir(record.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create audit dir: %w", err)
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}

	return os.WriteFile(filepath.Join(dir, "meta.json"), data, 0o644)
}

// Load reads an AuditRecord from disk.
func (s *Storage) Load(id string) (*AuditRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(filepath.Join(s.AuditDir(id), "meta.json"))
	if err != nil {
		return nil, fmt.Errorf("read meta for %q: %w", id, err)
	}

	var record AuditRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("unmarshal meta for %q: %w", id, err)
	}
	return &record, nil
}

// List returns all AuditRecords sorted by CreatedAt descending.
// Broken/missing meta files are silently skipped.
func (s *Storage) List() ([]*AuditRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("read storage dir: %w", err)
	}

	var records []*AuditRecord
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.baseDir, e.Name(), "meta.json"))
		if err != nil {
			continue
		}
		var r AuditRecord
		if err := json.Unmarshal(data, &r); err != nil {
			continue
		}
		records = append(records, &r)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	return records, nil
}

// Delete removes an audit directory and all its files from disk.
func (s *Storage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.RemoveAll(s.AuditDir(id))
}
