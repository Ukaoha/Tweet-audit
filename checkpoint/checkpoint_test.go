package checkpoint_test

import (
	"os"
	"path/filepath"
	"testing"

	"tweet-audit/audit"
	"tweet-audit/checkpoint"
)

func sampleVerdicts() []audit.Verdict {
	return []audit.Verdict{
		{TweetID: "111", Username: "alice", URL: "https://x.com/alice/status/111", Text: "hello", Flag: true, Reason: "rude"},
		{TweetID: "222", Username: "alice", URL: "https://x.com/alice/status/222", Text: "world", Flag: false, Reason: "fine"},
	}
}

func TestLoadNonExistent(t *testing.T) {
	s, err := checkpoint.Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if s.Len() != 0 {
		t.Fatalf("expected empty store, got %d", s.Len())
	}
}

func TestAddAndHas(t *testing.T) {
	s, _ := checkpoint.Load(filepath.Join(t.TempDir(), "cp.json"))
	v := sampleVerdicts()[0]
	s.Add(v)

	if !s.Has("111") {
		t.Error("expected Has(111) == true after Add")
	}
	if s.Has("999") {
		t.Error("expected Has(999) == false for unknown tweet")
	}
}

func TestGet(t *testing.T) {
	s, _ := checkpoint.Load(filepath.Join(t.TempDir(), "cp.json"))
	s.Add(sampleVerdicts()[0])

	got, ok := s.Get("111")
	if !ok {
		t.Fatal("expected Get to find verdict 111")
	}
	if got.Flag != true || got.Reason != "rude" {
		t.Errorf("unexpected verdict: %+v", got)
	}

	_, ok = s.Get("999")
	if ok {
		t.Error("expected Get to return false for unknown tweet")
	}
}

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cp.json")

	s, _ := checkpoint.Load(path)
	for _, v := range sampleVerdicts() {
		s.Add(v)
	}
	if err := s.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// File must exist after Save
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("checkpoint file not found after Save: %v", err)
	}

	// Reload and verify
	s2, err := checkpoint.Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if s2.Len() != 2 {
		t.Fatalf("expected 2 verdicts after reload, got %d", s2.Len())
	}
	for _, v := range sampleVerdicts() {
		if !s2.Has(v.TweetID) {
			t.Errorf("expected tweet %s to be present after reload", v.TweetID)
		}
	}
}

func TestSaveIsAtomic(t *testing.T) {
	// After Save, no .tmp file should remain
	dir := t.TempDir()
	path := filepath.Join(dir, "cp.json")

	s, _ := checkpoint.Load(path)
	s.Add(sampleVerdicts()[0])
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	tmp := path + ".tmp"
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Error("expected .tmp file to be cleaned up after Save")
	}
}

func TestAll(t *testing.T) {
	s, _ := checkpoint.Load(filepath.Join(t.TempDir(), "cp.json"))
	for _, v := range sampleVerdicts() {
		s.Add(v)
	}
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 from All(), got %d", len(all))
	}
}

func TestDeduplication(t *testing.T) {
	s, _ := checkpoint.Load(filepath.Join(t.TempDir(), "cp.json"))
	v := sampleVerdicts()[0]
	s.Add(v)
	v.Reason = "updated reason"
	s.Add(v) // overwrite same ID

	if s.Len() != 1 {
		t.Errorf("expected 1 entry after adding same ID twice, got %d", s.Len())
	}
	got, _ := s.Get("111")
	if got.Reason != "updated reason" {
		t.Errorf("expected updated reason, got: %s", got.Reason)
	}
}
