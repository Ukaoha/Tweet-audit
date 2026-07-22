// Package checkpoint persists processed tweet verdicts to disk so that
// interrupted audits can resume without re-calling the Gemini API.
package checkpoint

import (
	"encoding/json"
	"fmt"
	"os"

	"tweet-audit/audit"
)

// Store is an in-memory map of tweet ID → Verdict backed by a JSON file.
type Store struct {
	path    string
	verdicts map[string]audit.Verdict
}

// New initializes an empty Store at the given path.
func New(path string) *Store {
	return &Store{
		path:     path,
		verdicts: make(map[string]audit.Verdict),
	}
}

// Load reads an existing checkpoint file from path.
// If the file does not exist, an empty Store is returned (not an error).
func Load(path string) (*Store, error) {
	s := &Store{
		path:    path,
		verdicts: make(map[string]audit.Verdict),
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading checkpoint %s: %w", path, err)
	}

	var records []audit.Verdict
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("parsing checkpoint %s: %w", path, err)
	}

	for _, v := range records {
		s.verdicts[v.TweetID] = v
	}
	return s, nil
}

// Has returns true if the tweet has already been processed.
func (s *Store) Has(tweetID string) bool {
	_, ok := s.verdicts[tweetID]
	return ok
}

// Get returns the stored verdict for a tweet ID.
func (s *Store) Get(tweetID string) (audit.Verdict, bool) {
	v, ok := s.verdicts[tweetID]
	return v, ok
}

// Add records a new verdict in memory.
func (s *Store) Add(v audit.Verdict) {
	s.verdicts[v.TweetID] = v
}

// Save flushes all verdicts to the checkpoint file atomically.
// It writes to a temp file and renames so a crash mid-write doesn't corrupt data.
func (s *Store) Save() error {
	records := make([]audit.Verdict, 0, len(s.verdicts))
	for _, v := range s.verdicts {
		records = append(records, v)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling checkpoint: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("writing temp checkpoint: %w", err)
	}

	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("renaming checkpoint: %w", err)
	}
	return nil
}

// All returns a slice of every stored verdict.
func (s *Store) All() []audit.Verdict {
	result := make([]audit.Verdict, 0, len(s.verdicts))
	for _, v := range s.verdicts {
		result = append(result, v)
	}
	return result
}

// Len returns the number of stored verdicts.
func (s *Store) Len() int {
	return len(s.verdicts)
}
