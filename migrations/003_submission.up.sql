CREATE TABLE submissions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    form_id UUID NOT NULL REFERENCES forms(id),
    form_version_id UUID NOT NULL REFERENCES form_versions(id),
    data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_submissions_form_created
ON submissions(form_id, created_at DESC, id DESC);

CREATE INDEX idx_submissions_tenant_form_created
ON submissions(
    tenant_id,
    form_id,
    created_at DESC,
    id DESC
);