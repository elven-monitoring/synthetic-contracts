package job

import "testing"

func TestStreamName(t *testing.T) {
	got, err := StreamName(WorkerPoolAPI, PriorityHigh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "synthetic:jobs:api:high" {
		t.Fatalf("unexpected stream name: %q", got)
	}

	got, err = StreamName(WorkerPoolBrowser, PriorityNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "synthetic:jobs:browser:normal" {
		t.Fatalf("unexpected stream name: %q", got)
	}
}

func TestStreamName_InvalidInputs(t *testing.T) {
	if _, err := StreamName("nope", PriorityHigh); err == nil {
		t.Fatalf("expected error for invalid pool")
	}
	if _, err := StreamName(WorkerPoolAPI, "nope"); err == nil {
		t.Fatalf("expected error for invalid priority")
	}
}

