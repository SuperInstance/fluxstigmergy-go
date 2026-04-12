package fluxstigmergy

import (
	"testing"
	"time"
)

func TestDepositAndRead(t *testing.T) {
	e := NewEnv()
	idx := e.Deposit(1, "food", "at nest", InfoTrace, 500)
	if idx != 0 {
		t.Fatalf("expected index 0, got %d", idx)
	}
	tr := e.Read("food")
	if tr == nil || tr.Value != "at nest" || tr.AuthorId != 1 {
		t.Fatalf("unexpected trace: %+v", tr)
	}
}

func TestReadNonexistentNil(t *testing.T) {
	e := NewEnv()
	if e.Read("nope") != nil {
		t.Fatal("expected nil")
	}
}

func TestReadAllPrefix(t *testing.T) {
	e := NewEnv()
	e.Deposit(1, "path:a", "x", InfoTrace, 100)
	e.Deposit(1, "path:b", "y", InfoTrace, 100)
	e.Deposit(1, "other:c", "z", InfoTrace, 100)
	results := e.ReadAll("path:", 10)
	if len(results) != 2 {
		t.Fatalf("expected 2, got %d", len(results))
	}
	results = e.ReadAll("path:", 1)
	if len(results) != 1 {
		t.Fatalf("expected 1, got %d", len(results))
	}
}

func TestModifyExisting(t *testing.T) {
	e := NewEnv()
	e.Deposit(1, "key1", "old", InfoTrace, 200)
	ok := e.Modify(1, "key1", "new", 50)
	if !ok {
		t.Fatal("expected true")
	}
	tr := e.Read("key1")
	if tr.Value != "new" || tr.Strength != 250 {
		t.Fatalf("unexpected: value=%s strength=%d", tr.Value, tr.Strength)
	}
}

func TestModifyNonexistentFalse(t *testing.T) {
	e := NewEnv()
	if e.Modify(1, "nope", "x", 10) {
		t.Fatal("expected false")
	}
}

func TestEraseByAuthorTrue(t *testing.T) {
	e := NewEnv()
	e.Deposit(1, "del", "x", InfoTrace, 100)
	if !e.Erase(1, "del") {
		t.Fatal("expected true")
	}
	if e.Read("del") != nil {
		t.Fatal("trace should be gone")
	}
}

func TestEraseByWrongAuthorFalse(t *testing.T) {
	e := NewEnv()
	e.Deposit(1, "del", "x", InfoTrace, 100)
	if e.Erase(2, "del") {
		t.Fatal("expected false")
	}
}

func TestDecayReducesStrength(t *testing.T) {
	e := NewEnv()
	e.Deposit(1, "d1", "v", InfoTrace, 500)
	// Manually backdate so decay actually reduces
	e.traces[0].Timestamp = time.Now().Add(-2 * time.Hour)
	removed := e.Decay(1*time.Hour, 0, 0)
	if removed != 0 {
		t.Fatalf("expected 0 removed, got %d", removed)
	}
	if e.traces[0].Strength >= 500 {
		t.Fatalf("strength should have decreased, got %d", e.traces[0].Strength)
	}
}

func TestGarbageCollectWeak(t *testing.T) {
	e := NewEnv()
	e.Deposit(1, "weak", "v", InfoTrace, 5)
	e.traces[0].Timestamp = time.Now().Add(-10 * time.Hour)
	removed := e.Decay(1*time.Hour, 0, 10)
	if removed != 1 {
		t.Fatalf("expected 1 removed, got %d", removed)
	}
	if len(e.traces) != 0 {
		t.Fatal("trace should be gone")
	}
}

func TestByAuthorFilter(t *testing.T) {
	e := NewEnv()
	e.Deposit(1, "a", "x", InfoTrace, 100)
	e.Deposit(2, "b", "y", InfoTrace, 100)
	e.Deposit(1, "c", "z", InfoTrace, 100)
	results := e.ByAuthor(1)
	if len(results) != 2 {
		t.Fatalf("expected 2, got %d", len(results))
	}
}

func TestByTypeFilter(t *testing.T) {
	e := NewEnv()
	e.Deposit(1, "a", "x", InfoTrace, 100)
	e.Deposit(1, "b", "y", ClaimTrace, 100)
	e.Deposit(1, "c", "z", InfoTrace, 100)
	results := e.ByType(InfoTrace)
	if len(results) != 2 {
		t.Fatalf("expected 2, got %d", len(results))
	}
}

func TestStrongestN(t *testing.T) {
	e := NewEnv()
	e.Deposit(1, "a", "x", InfoTrace, 100)
	e.Deposit(1, "b", "y", InfoTrace, 500)
	e.Deposit(1, "c", "z", InfoTrace, 300)
	results := e.Strongest(2)
	if len(results) != 2 || results[0].Strength != 500 || results[1].Strength != 300 {
		t.Fatalf("unexpected strongest: %v", results)
	}
}

func TestOldestN(t *testing.T) {
	e := NewEnv()
	e.Deposit(1, "a", "x", InfoTrace, 100)
	now := time.Now()
	e.traces[0].Timestamp = now.Add(-2 * time.Hour)
	e.Deposit(1, "b", "y", InfoTrace, 100)
	e.traces[1].Timestamp = now.Add(-1 * time.Hour)
	e.Deposit(1, "c", "z", InfoTrace, 100)
	e.traces[2].Timestamp = now
	results := e.Oldest(2)
	if len(results) != 2 {
		t.Fatalf("expected 2, got %d", len(results))
	}
	if !results[0].Timestamp.Before(results[1].Timestamp) {
		t.Fatal("first should be older")
	}
}

func TestStatsAccurate(t *testing.T) {
	e := NewEnv()
	e.Deposit(1, "a", "x", InfoTrace, 200)
	e.Deposit(2, "b", "y", ClaimTrace, 300)
	s := e.Stats(0)
	if s.Total != 2 || s.Active != 2 || s.UniqueAuthors != 2 {
		t.Fatalf("unexpected stats: %+v", s)
	}
	if s.ByType[InfoTrace] != 1 || s.ByType[ClaimTrace] != 1 {
		t.Fatalf("byType wrong: %+v", s.ByType)
	}
}

func TestReadIncrementsCount(t *testing.T) {
	e := NewEnv()
	e.Deposit(1, "cnt", "v", InfoTrace, 100)
	e.Read("cnt")
	e.Read("cnt")
	tr := e.Read("cnt")
	if tr.Reads != 3 {
		t.Fatalf("expected 3 reads, got %d", tr.Reads)
	}
}

func TestModifyCapsStrength(t *testing.T) {
	e := NewEnv()
	e.Deposit(1, "cap", "v", InfoTrace, 900)
	e.Modify(1, "cap", "v2", 500)
	tr := e.Read("cap")
	if tr.Strength != 1000 {
		t.Fatalf("expected capped 1000, got %d", tr.Strength)
	}
}

func TestReadAllIncrementsReads(t *testing.T) {
	e := NewEnv()
	e.Deposit(1, "pfx:1", "a", InfoTrace, 100)
	e.Deposit(1, "pfx:2", "b", InfoTrace, 100)
	e.ReadAll("pfx:", 10)
	tr := e.Read("pfx:1")
	if tr.Reads != 2 {
		t.Fatalf("expected 2, got %d", tr.Reads)
	}
}
