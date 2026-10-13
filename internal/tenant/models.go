package tenant

import "time"

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type TenantList struct {
	Tenants []Tenant `json:"tenants"`
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
	Total   int      `json:"total"`
}