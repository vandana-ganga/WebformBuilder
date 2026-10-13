package submission

import "time"

// Submission represents a submission stored in PostgreSQL.
type Submission struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	FormID        string    `json:"form_id"`
	FormVersionID string    `json:"form_version_id"`
	Data          any       `json:"data"`
	CreatedAt     time.Time `json:"created_at"`
}

// to NATS JetStream.
type SubmissionMessage struct {
	SubmissionID  string    `json:"submission_id"`
	TenantID      string    `json:"tenant_id"`
	FormID        string    `json:"form_id"`
	FormVersionID string    `json:"form_version_id"`
	Data          any       `json:"data"`
	CreatedAt     time.Time `json:"created_at"`
}