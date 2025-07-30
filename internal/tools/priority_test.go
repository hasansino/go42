package tools_test

import (
	"testing"

	"github.com/hasansino/go42/internal/tools"
)

func TestPriorityQueue_EnqueueDequeue(t *testing.T) {
	pq := tools.NewPriorityQueue[int]()
	pq.Enqueue(2, 20)
	pq.Enqueue(1, 10)
	pq.Enqueue(3, 30)

	v, ok := pq.Dequeue()
	if !ok || v != 10 {
		t.Errorf("expected 10, got %v", v)
	}
	v, ok = pq.Dequeue()
	if !ok || v != 20 {
		t.Errorf("expected 20, got %v", v)
	}
	v, ok = pq.Dequeue()
	if !ok || v != 30 {
		t.Errorf("expected 30, got %v", v)
	}
	_, ok = pq.Dequeue()
	if ok {
		t.Error("expected queue to be empty")
	}
}

func TestPriorityQueue_IsEmpty(t *testing.T) {
	pq := tools.NewPriorityQueue[string]()
	if !pq.IsEmpty() {
		t.Error("expected queue to be empty initially")
	}
	pq.Enqueue(1, "a")
	if pq.IsEmpty() {
		t.Error("expected queue to not be empty after enqueue")
	}
	pq.Dequeue()
	if !pq.IsEmpty() {
		t.Error("expected queue to be empty after dequeue")
	}
}

func TestPriorityQueue_MultipleSamePriority(t *testing.T) {
	pq := tools.NewPriorityQueue[int]()
	pq.Enqueue(1, 10)
	pq.Enqueue(1, 20)
	pq.Enqueue(1, 30)

	v, _ := pq.Dequeue()
	if v != 10 {
		t.Errorf("expected 10, got %v", v)
	}
	v, _ = pq.Dequeue()
	if v != 20 {
		t.Errorf("expected 20, got %v", v)
	}
	v, _ = pq.Dequeue()
	if v != 30 {
		t.Errorf("expected 30, got %v", v)
	}
}

func TestPriorityQueue_Extract(t *testing.T) {
	pq := tools.NewPriorityQueue[string]()
	if len(pq.Extract()) != 0 {
		t.Error("expected empty slice from empty queue")
	}
	pq.Enqueue(2, "b")
	pq.Enqueue(1, "a")
	pq.Enqueue(3, "c")
	pq.Enqueue(2, "bb")
	pq.Enqueue(1, "aa")
	result := pq.Extract()
	// Should be: ["a", "aa", "b", "bb", "c"]
	expected := []string{"a", "aa", "b", "bb", "c"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(result))
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("at %d: expected %q, got %q", i, v, result[i])
		}
	}
}
