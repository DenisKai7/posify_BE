-- ==============================================================================
-- 003_SEED_DATA.SQL (WITH PASSWORD HASH FOR GO BACKEND AUTH)
-- Hash di bawah adalah bcrypt dari: password123
-- ==============================================================================

-- 0. ALTER TABLE (Pastikan kolom password_hash ada di profiles jika belum dibuat)
ALTER TABLE public.profiles ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);

-- 1. Demo Tenant
INSERT INTO public.tenants (id, name)
VALUES ('a0000000-0000-0000-0000-000000000001', 'Toko Posify Demo')
ON CONFLICT (id) DO NOTHING;

-- 2. Seed Users dengan Password Hash (password: password123)
-- Owner
INSERT INTO public.profiles (id, tenant_id, email, full_name, role, password_hash)
VALUES (
  'b0000000-0000-0000-0000-000000000001',
  'a0000000-0000-0000-0000-000000000001',
  'owner@posify.com',
  'Owner Demo',
  'OWNER',
  '$2a$10$9TG7Oaa.FomIAqPdPzsb7ujpZs0Yk1u0rgGPIe7Th0DGR1LF4B7ZG'
) ON CONFLICT (email) DO UPDATE 
SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role;

-- Manager
INSERT INTO public.profiles (id, tenant_id, email, full_name, role, password_hash)
VALUES (
  'b0000000-0000-0000-0000-000000000002',
  'a0000000-0000-0000-0000-000000000001',
  'manager@posify.com',
  'Manager Demo',
  'MANAGER',
  '$2a$10$9TG7Oaa.FomIAqPdPzsb7ujpZs0Yk1u0rgGPIe7Th0DGR1LF4B7ZG'
) ON CONFLICT (email) DO UPDATE 
SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role;

-- Cashier
INSERT INTO public.profiles (id, tenant_id, email, full_name, role, password_hash)
VALUES (
  'b0000000-0000-0000-0000-000000000003',
  'a0000000-0000-0000-0000-000000000001',
  'cashier@posify.com',
  'Cashier Demo',
  'CASHIER',
  '$2a$10$9TG7Oaa.FomIAqPdPzsb7ujpZs0Yk1u0rgGPIe7Th0DGR1LF4B7ZG'
) ON CONFLICT (email) DO UPDATE 
SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role;

-- 3. Sample Products
INSERT INTO public.products (id, tenant_id, sku, name, price, stock, category)
VALUES
  ('d0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'MNM-001', 'Es Teh Manis', 5000, 100, 'Minuman'),
  ('d0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'MNM-002', 'Kopi Susu', 15000, 50, 'Minuman'),
  ('d0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001', 'MNM-003', 'Jus Jeruk', 12000, 30, 'Minuman'),
  ('d0000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001', 'MKN-001', 'Nasi Goreng', 20000, 25, 'Makanan'),
  ('d0000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001', 'MKN-002', 'Mie Goreng', 18000, 30, 'Makanan'),
  ('d0000000-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000001', 'MKN-003', 'Ayam Goreng', 25000, 20, 'Makanan'),
  ('d0000000-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000001', 'MKN-004', 'Sate Ayam', 22000, 15, 'Makanan'),
  ('d0000000-0000-0000-0000-000000000008', 'a0000000-0000-0000-0000-000000000001', 'MNM-004', 'Air Mineral', 3000, 200, 'Minuman')
ON CONFLICT (sku) DO NOTHING;