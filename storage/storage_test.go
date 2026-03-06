package storage

import (
	"reflect"
	"testing"
)

func TestMemoryStoreEvictOldest(t *testing.T) {
	s := NewMemory(3)
	if err := s.Store("a", "b", "c", "d", "e"); err != nil {
		t.Fatalf("store failed: %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	want := []string{"c", "d", "e"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected history: got=%v want=%v", got, want)
	}
}

func TestRingStoreEvictOldest(t *testing.T) {
	s := NewRing(3)
	if err := s.Store("a", "b", "c", "d", "e"); err != nil {
		t.Fatalf("store failed: %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	want := []string{"c", "d", "e"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected history: got=%v want=%v", got, want)
	}
}

func TestStorageClear(t *testing.T) {
	tests := []struct {
		name string
		new  func() Storage
	}{
		{name: "memory", new: func() Storage { return NewMemory(4) }},
		{name: "ring", new: func() Storage { return NewRing(4) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.new()
			if err := s.Store("x", "y"); err != nil {
				t.Fatalf("store failed: %v", err)
			}
			if err := s.Clear(); err != nil {
				t.Fatalf("clear failed: %v", err)
			}

			got, err := s.Load()
			if err != nil {
				t.Fatalf("load failed: %v", err)
			}
			if len(got) != 0 {
				t.Fatalf("expected empty history after clear, got=%v", got)
			}
		})
	}
}

func TestStorageLoadReturnsCopy(t *testing.T) {
	tests := []struct {
		name string
		new  func() Storage
	}{
		{name: "memory", new: func() Storage { return NewMemory(4) }},
		{name: "ring", new: func() Storage { return NewRing(4) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.new()
			if err := s.Store("a", "b"); err != nil {
				t.Fatalf("store failed: %v", err)
			}

			first, err := s.Load()
			if err != nil {
				t.Fatalf("first load failed: %v", err)
			}
			if len(first) != 2 {
				t.Fatalf("unexpected first load length: %d", len(first))
			}
			first[0] = "changed"

			second, err := s.Load()
			if err != nil {
				t.Fatalf("second load failed: %v", err)
			}
			if len(second) != 2 {
				t.Fatalf("unexpected second load length: %d", len(second))
			}
			if second[0] != "a" {
				t.Fatalf("storage should return copy, got=%v", second)
			}
		})
	}
}
