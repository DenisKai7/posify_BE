-- 005_shifts.sql

DO $$ BEGIN
    CREATE TYPE shift_status AS ENUM ('OPEN', 'CLOSED');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS shifts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES profiles(id),
    initial_cash DECIMAL(12,2) NOT NULL DEFAULT 0,
    actual_cash DECIMAL(12,2),
    expected_cash DECIMAL(12,2),
    difference DECIMAL(12,2),
    total_sales DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_transactions INT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMPTZ,
    status shift_status NOT NULL DEFAULT 'OPEN'
);

CREATE INDEX IF NOT EXISTS idx_shifts_tenant_user ON shifts(tenant_id, user_id, status);
