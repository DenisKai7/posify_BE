-- 004_fix_seed_profiles.sql
-- Fix: reassign seed profiles to "Toko Posify Demo" tenant and ensure password hashes exist

-- Update profiles to point to the correct demo tenant
UPDATE profiles
SET tenant_id = 'a0000000-0000-0000-0000-000000000001'
WHERE email IN ('owner@posify.com', 'manager@posify.com', 'cashier@posify.com')
  AND tenant_id = '00000000-0000-0000-0000-000000000001';

-- Ensure password hashes are set (bcrypt of "password123")
UPDATE profiles
SET password_hash = '$2a$10$9TG7Oaa.FomIAqPdPzsb7ujpZs0Yk1u0rgGPIe7Th0DGR1LF4B7ZG'
WHERE email IN ('owner@posify.com', 'manager@posify.com', 'cashier@posify.com')
  AND (password_hash IS NULL OR password_hash = '');
