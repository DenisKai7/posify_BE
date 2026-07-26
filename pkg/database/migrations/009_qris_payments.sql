-- 009_qris_payments.sql

DO $$ BEGIN
    CREATE TYPE qris_status AS ENUM ('PENDING', 'PAID', 'EXPIRED', 'FAILED');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS qris_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    order_id VARCHAR(100) UNIQUE NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    qr_string TEXT NOT NULL,
    qr_url TEXT,
    status qris_status NOT NULL DEFAULT 'PENDING',
    payment_method VARCHAR(50),
    paid_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_qris_order ON qris_payments(order_id);
CREATE INDEX IF NOT EXISTS idx_qris_status ON qris_payments(status, expires_at);
