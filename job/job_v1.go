package job

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const JobSchemaVersionV1 = 1

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
)

type ScriptRef struct {
	ScriptID uuid.UUID `json:"script_id"`
	Version  int       `json:"version"`
	Sha256   string    `json:"sha256,omitempty"`
}

type Limits struct {
	TimeoutSeconds    int `json:"timeout_seconds,omitempty"`
	MaxVUs            int `json:"max_vus,omitempty"`
	MaxDurationSeconds int `json:"max_duration_seconds,omitempty"`

	// Resource hints for sandboxed execution modes.
	MemoryMB      int `json:"memory_mb,omitempty"`
	CPUMillicores int `json:"cpu_millicores,omitempty"`

	MaxLogBytes      int64 `json:"max_log_bytes,omitempty"`
	MaxArtifactBytes int64 `json:"max_artifact_bytes,omitempty"`
}

type ExecutionJobV1 struct {
	JobSchemaVersion int `json:"job_schema_version"`

	JobID       uuid.UUID `json:"job_id"`
	ExecutionID uuid.UUID `json:"execution_id"`

	TenantID string `json:"tenant_id"`

	// W3C trace context (optional). Propagate across job boundaries.
	TraceParent string `json:"traceparent,omitempty"`

	IdempotencyKey string `json:"idempotency_key"`

	TriggerType string     `json:"trigger_type"` // manual|scheduled|api|ci-cd|immediate
	ScheduledFor *time.Time `json:"scheduled_for,omitempty"`

	Priority Priority `json:"priority"`

	Attempt     int `json:"attempt"`
	MaxAttempts int `json:"max_attempts"`

	NextVisibleAt *time.Time `json:"next_visible_at,omitempty"`

	ScriptRef   ScriptRef   `json:"script_ref"`
	ScenarioID  *uuid.UUID  `json:"scenario_id,omitempty"`
	Config      ExecutionConfigV1 `json:"config"`
	Limits      Limits      `json:"limits,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

type ExecutionConfigV1 struct {
	UserID      string `json:"user_id,omitempty"`
	Region      string `json:"region,omitempty"`
	Environment EnvironmentConfig `json:"environment"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	LoadConfig  LoadConfig `json:"load_config"`
	OverrideVars map[string]string `json:"override_vars,omitempty"`
	DryRun      bool `json:"dry_run,omitempty"`

	ScriptType     string `json:"script_type,omitempty"`
	BrowserEnabled bool   `json:"browser_enabled,omitempty"`
}

type LoadConfig struct {
	VUs        int     `json:"vus,omitempty"`
	Duration   string  `json:"duration,omitempty"`
	Stages     []Stage `json:"stages,omitempty"`
	Iterations int     `json:"iterations,omitempty"`
	Rate       int     `json:"rate,omitempty"`
}

type Stage struct {
	Duration string `json:"duration"`
	Target   int    `json:"target"`
}

type EnvironmentConfig struct {
	BaseURL    string            `json:"base_url"`
	Headers    map[string]string `json:"headers,omitempty"`
	AuthConfig AuthConfig        `json:"auth_config,omitempty"`
	Timeout    int               `json:"timeout,omitempty"` // seconds
}

type AuthConfig struct {
	Type         string `json:"type,omitempty"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	Token        string `json:"token,omitempty"`
	APIKey       string `json:"api_key,omitempty"`
	APIKeyHeader string `json:"api_key_header,omitempty"`
}

func (j *ExecutionJobV1) ValidateBasic() error {
	if j.JobSchemaVersion != JobSchemaVersionV1 {
		return fmt.Errorf("unsupported job_schema_version: %d", j.JobSchemaVersion)
	}
	if j.JobID == uuid.Nil {
		return fmt.Errorf("job_id is required")
	}
	if j.ExecutionID == uuid.Nil {
		return fmt.Errorf("execution_id is required")
	}
	if j.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if j.IdempotencyKey == "" {
		return fmt.Errorf("idempotency_key is required")
	}
	if j.ScriptRef.ScriptID == uuid.Nil {
		return fmt.Errorf("script_ref.script_id is required")
	}
	if j.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}
	return nil
}

// ComputeIdempotencyKey returns a stable key for end-to-end deduplication.
//
// Default formula (per plan): sha256(tenant_id + script_ref + scheduled_for + trigger_type).
// Use scheduledForUTC with a stable format (RFC3339Nano) to avoid timezone drift.
func ComputeIdempotencyKey(tenantID string, scriptRef ScriptRef, scheduledForUTC time.Time, triggerType string) string {
	h := sha256.New()
	_, _ = h.Write([]byte(tenantID))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(scriptRef.ScriptID.String()))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(fmt.Sprintf("%d", scriptRef.Version)))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(scheduledForUTC.UTC().Format(time.RFC3339Nano)))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(triggerType))
	return hex.EncodeToString(h.Sum(nil))
}

// DeterministicExecutionID derives a UUID from an idempotency key.
// This makes retries/duplicates converge on the same execution_id.
func DeterministicExecutionID(idempotencyKey string) uuid.UUID {
	// NewSHA1 is a SHA1-based UUID (v5 semantics).
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(idempotencyKey))
}

