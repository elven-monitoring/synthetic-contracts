package job

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestComputeIdempotencyKey_DeterministicAndSensitiveToInputs(t *testing.T) {
	tenantID := "tenant-a"
	scriptID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	scriptRef := ScriptRef{ScriptID: scriptID, Version: 7, Sha256: "abc"}
	scheduledFor := time.Date(2026, 2, 6, 12, 34, 56, 789000000, time.FixedZone("X", -3*3600))

	k1 := ComputeIdempotencyKey(tenantID, scriptRef, scheduledFor, "scheduled")
	k2 := ComputeIdempotencyKey(tenantID, scriptRef, scheduledFor, "scheduled")
	if k1 == "" || k2 == "" {
		t.Fatalf("expected non-empty keys: %q %q", k1, k2)
	}
	if k1 != k2 {
		t.Fatalf("expected deterministic keys, got %q vs %q", k1, k2)
	}

	// Changing any field should change the key.
	if k := ComputeIdempotencyKey("tenant-b", scriptRef, scheduledFor, "scheduled"); k == k1 {
		t.Fatalf("expected tenant change to change key")
	}
	if k := ComputeIdempotencyKey(tenantID, ScriptRef{ScriptID: uuid.New(), Version: 7}, scheduledFor, "scheduled"); k == k1 {
		t.Fatalf("expected script_id change to change key")
	}
	if k := ComputeIdempotencyKey(tenantID, ScriptRef{ScriptID: scriptID, Version: 8}, scheduledFor, "scheduled"); k == k1 {
		t.Fatalf("expected version change to change key")
	}
	if k := ComputeIdempotencyKey(tenantID, scriptRef, scheduledFor.Add(1*time.Second), "scheduled"); k == k1 {
		t.Fatalf("expected scheduled_for change to change key")
	}
	if k := ComputeIdempotencyKey(tenantID, scriptRef, scheduledFor, "manual"); k == k1 {
		t.Fatalf("expected trigger_type change to change key")
	}
}

func TestDeterministicExecutionID_Stable(t *testing.T) {
	key := "deadbeef"
	a := DeterministicExecutionID(key)
	b := DeterministicExecutionID(key)
	if a == uuid.Nil || b == uuid.Nil {
		t.Fatalf("expected non-nil ids")
	}
	if a != b {
		t.Fatalf("expected stable deterministic UUID, got %s vs %s", a, b)
	}
}

func TestExecutionJobV1_ValidateBasic(t *testing.T) {
	now := time.Now().UTC()
	j := &ExecutionJobV1{
		JobSchemaVersion: JobSchemaVersionV1,
		JobID:            uuid.New(),
		ExecutionID:      uuid.New(),
		TenantID:         "t",
		IdempotencyKey:   "k",
		TriggerType:      "manual",
		Priority:         PriorityHigh,
		Attempt:          1,
		MaxAttempts:      10,
		ScriptRef:        ScriptRef{ScriptID: uuid.New(), Version: 1},
		CreatedAt:        now,
	}

	if err := j.ValidateBasic(); err != nil {
		t.Fatalf("expected valid job, got err: %v", err)
	}

	j.JobSchemaVersion = 999
	if err := j.ValidateBasic(); err == nil {
		t.Fatalf("expected validation error for unsupported schema version")
	}
}

