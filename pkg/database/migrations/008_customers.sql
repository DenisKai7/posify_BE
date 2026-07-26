-- 008_customers.sql

DO $$ BEGIN
    CREATE TYPE customer_tier AS ENUM ('SILVER', 'GOLD', 'PLATINUM');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(100),
    points INT NOT NULL DEFAULT 0,
    tier customer_tier NOT NULL DEFAULT 'SILVER',
    total_spent DECIMAL(14,2) NOT NULL DEFAULT 0,
    visit_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_phone_per_tenant UNIQUE (tenant_id, phone)
);

CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(tenant_id, phone);
CREATE INDEX IF NOT EXISTS idx_customers_name ON customers(tenant_id, name);
