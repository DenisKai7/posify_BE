-- 011_payment_gateways_tiers.sql

-- Payment Gateway configuration per tenant
CREATE TABLE IF NOT EXISTS payment_gateways (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL DEFAULT 'MIDTRANS',
    merchant_id VARCHAR(100),
    client_key VARCHAR(255),
    server_key VARCHAR(255),
    is_production BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    webhook_url VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_provider_per_tenant UNIQUE (tenant_id, provider)
);

-- Membership Tiers
CREATE TABLE IF NOT EXISTS membership_tiers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    min_spend DECIMAL(14,2) NOT NULL DEFAULT 0,
    multiplier_points DECIMAL(4,2) NOT NULL DEFAULT 1.0,
    discount_percentage DECIMAL(5,2) NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_tier_name_per_tenant UNIQUE (tenant_id, name)
);

-- Add tier_id to customers if not exists
DO $$ BEGIN
    ALTER TABLE customers ADD COLUMN tier_id UUID REFERENCES membership_tiers(id);
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

-- Enhance discounts table
DO $$ BEGIN
    ALTER TABLE discounts ADD COLUMN applies_to_tier UUID REFERENCES membership_tiers(id);
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE discounts ADD COLUMN buy_x INT;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE discounts ADD COLUMN get_y INT;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE discounts ADD COLUMN applies_to_product_id UUID REFERENCES products(id);
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

-- Seed default tiers
INSERT INTO membership_tiers (tenant_id, name, min_spend, multiplier_points, discount_percentage, sort_order)
SELECT t.id, tier.name, tier.min_spend, tier.multiplier, tier.disc, tier.ord
FROM tenants t
CROSS JOIN (VALUES
    ('Bronze', 0, 1.0, 0, 1),
    ('Silver', 500000, 1.5, 2, 2),
    ('Gold', 2000000, 2.0, 5, 3),
    ('Platinum', 10000000, 3.0, 10, 4)
) AS tier(name, min_spend, multiplier, disc, ord)
ON CONFLICT (tenant_id, name) DO NOTHING;
