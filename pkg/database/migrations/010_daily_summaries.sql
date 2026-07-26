-- 010_daily_summaries.sql

CREATE TABLE IF NOT EXISTS daily_summaries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    summary_date DATE NOT NULL,
    total_revenue DECIMAL(14,2) NOT NULL DEFAULT 0,
    total_transactions INT NOT NULL DEFAULT 0,
    total_items_sold INT NOT NULL DEFAULT 0,
    cash_revenue DECIMAL(14,2) NOT NULL DEFAULT 0,
    qris_revenue DECIMAL(14,2) NOT NULL DEFAULT 0,
    top_product_name VARCHAR(255),
    top_product_qty INT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_summary_per_tenant_date UNIQUE (tenant_id, summary_date)
);

CREATE INDEX IF NOT EXISTS idx_daily_summaries_date ON daily_summaries(tenant_id, summary_date DESC);
