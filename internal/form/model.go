package form

import "time"

type Form struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FormDetail struct {
	Form
	VersionCount int      `json:"version_count"`
	VersionIDs   []string `json:"version_ids"`
}

type FormList struct {
	Forms []Form `json:"forms"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Total int    `json:"total"`
}

type FormVersionList struct {
	Versions []FormVersion `json:"versions"`
	Page     int           `json:"page"`
	Limit    int           `json:"limit"`
	Total    int           `json:"total"`
}

type FormVersion struct {
	ID            string     `json:"id"`
	FormID        string     `json:"form_id"`
	VersionNumber int        `json:"version_number"`
	Schema        any        `json:"schema"`
	CreatedAt     time.Time  `json:"created_at"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
}