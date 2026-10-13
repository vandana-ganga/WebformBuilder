ALTER TABLE tenants
    ADD COLUMN email TEXT,
    ADD COLUMN password_hash TEXT;

CREATE UNIQUE INDEX idx_tenants_email
ON tenants (LOWER(email));
