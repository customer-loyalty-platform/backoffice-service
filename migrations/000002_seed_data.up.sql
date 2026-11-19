-- Migration: 000002_seed_data.up.sql
-- Description: Seed initial admin user, tiers, sample partners, earning rules, sample members, and inventories

-- 1. Seed Admin User
-- Password: "password123" hashed with bcrypt (cost 10)
INSERT INTO schema_backoffice.admin_users (
    id, username, email, password_hash, full_name, role, permissions, is_active
) VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'superadmin',
    'admin@loyalty.local',
    '$2a$10$AVe9upDbXy8JkDbAFMWkjOJOysViI7olvJxHBIK9TXTUCbM.aO0Ru',
    'Super Administrator',
    'SUPER_ADMIN',
    '["ALL_ACCESS"]'::jsonb,
    TRUE
) ON CONFLICT (username) DO NOTHING;

-- 2. Seed Tier Definitions
INSERT INTO schema_backoffice.tier_definitions (
    id, tier_code, tier_name, min_points, max_points, multiplier, color_hex, perks, sort_order
) VALUES 
(
    'b0000000-0000-0000-0000-000000000001',
    'BRONZE',
    'Bronze Tier',
    0,
    999,
    1.00,
    '#CD7F32',
    '[{"code": "WELCOME_GIFT", "name": "Welcome Bonus", "description": "100 bonus points upon signup"}]'::jsonb,
    1
),
(
    'b0000000-0000-0000-0000-000000000002',
    'SILVER',
    'Silver Tier',
    1000,
    4999,
    1.25,
    '#C0C0C0',
    '[{"code": "FREE_DELIVERY", "name": "Free Delivery", "description": "Free standard shipping on all orders"}, {"code": "BIRTHDAY_BONUS", "name": "Birthday Bonus", "description": "2x points during birthday month"}]'::jsonb,
    2
),
(
    'b0000000-0000-0000-0000-000000000003',
    'GOLD',
    'Gold Tier',
    5000,
    19999,
    1.50,
    '#FFD700',
    '[{"code": "VIP_SUPPORT", "name": "Priority Support", "description": "Dedicated CS helpline"}, {"code": "BONUS_MULTIPLIER", "name": "1.5x Multiplier", "description": "1.5x points on every transaction"}]'::jsonb,
    3
),
(
    'b0000000-0000-0000-0000-000000000004',
    'PLATINUM',
    'Platinum Tier',
    20000,
    NULL,
    2.00,
    '#E5E4E2',
    '[{"code": "ALL_PERKS", "name": "All Perks", "description": "Access to all exclusive events and concierge service"}, {"code": "AIRPORT_LOUNGE", "name": "Airport Lounge", "description": "Free airport lounge pass 4x a year"}]'::jsonb,
    4
) ON CONFLICT (tier_code) DO NOTHING;

-- 3. Seed Sample Partners
INSERT INTO schema_backoffice.partners (
    id, partner_code, name, type, status, contact_email, settlement_rate
) VALUES 
(
    'c0000000-0000-0000-0000-000000000001',
    'PARTNER_STARBUCKS',
    'Starbucks Indonesia',
    'F&B',
    'ACTIVE',
    'partners@starbucks.co.id',
    0.0100
),
(
    'c0000000-0000-0000-0000-000000000002',
    'PARTNER_SHELL',
    'Shell Mobility',
    'FUEL',
    'ACTIVE',
    'b2b@shell.co.id',
    0.0120
) ON CONFLICT (partner_code) DO NOTHING;

-- 4. Seed Dynamic Earning Rule
INSERT INTO schema_backoffice.dynamic_rules (
    id, rule_code, rule_name, description, rule_type, status, priority, conditions_json, action_json
) VALUES (
    'd0000000-0000-0000-0000-000000000001',
    'RULE_BASE_FNB_1X',
    'Base F&B Earning Rule',
    'Standard earning rate 1 point per 10,000 IDR spent on F&B',
    'SPEND_BASED',
    'ACTIVE',
    10,
    '{"and": [{"field": "transaction.category", "operator": "EQUALS", "value": "F&B"}, {"field": "transaction.amount", "operator": "GREATER_THAN", "value": 0}]}'::jsonb,
    '{"type": "FIXED_RATIO", "points_per_unit": 1, "unit_amount": 10000}'::jsonb
) ON CONFLICT (rule_code) DO NOTHING;

-- 5. Seed Sample Members (matching actual schema_ledger.members columns)
INSERT INTO schema_ledger.members (
    id, account_number, current_tier_id, status, created_at, updated_at
) VALUES 
(
    'e0000000-0000-0000-0000-000000000001',
    'MBR-2024-0001',
    'b0000000-0000-0000-0000-000000000003', -- GOLD
    'ACTIVE',
    NOW() - INTERVAL '60 days',
    NOW()
),
(
    'e0000000-0000-0000-0000-000000000002',
    'MBR-2024-0002',
    'b0000000-0000-0000-0000-000000000002', -- SILVER
    'ACTIVE',
    NOW() - INTERVAL '30 days',
    NOW()
),
(
    'e0000000-0000-0000-0000-000000000003',
    'MBR-2024-0003',
    'b0000000-0000-0000-0000-000000000001', -- BRONZE
    'SUSPENDED',
    NOW() - INTERVAL '10 days',
    NOW()
) ON CONFLICT (account_number) DO NOTHING;

-- 6. Seed Sample Balances (matching actual schema_ledger.member_balances columns)
INSERT INTO schema_ledger.member_balances (
    member_id, total_points, locked_points, expiring_points, version, updated_at
) VALUES 
(
    'e0000000-0000-0000-0000-000000000001',
    9500.00,
    0.00,
    500.00,
    1,
    NOW()
),
(
    'e0000000-0000-0000-0000-000000000002',
    3000.00,
    0.00,
    200.00,
    1,
    NOW()
),
(
    'e0000000-0000-0000-0000-000000000003',
    800.00,
    800.00,
    0.00,
    1,
    NOW()
) ON CONFLICT (member_id) DO NOTHING;

-- 7. Seed Sample Voucher Batches (matching actual schema_burning.voucher_batches columns)
INSERT INTO schema_burning.voucher_batches (
    id, batch_code, title, points_cost, total_stock, available_stock, created_at
) VALUES 
(
    'f0000000-0000-0000-0000-000000000001',
    'VB-SBX-50K',
    'Starbucks IDR 50,000 E-Voucher',
    5000,
    1000,
    950,
    NOW()
),
(
    'f0000000-0000-0000-0000-000000000002',
    'VB-SHL-100K',
    'Shell Fuel Voucher IDR 100,000',
    10000,
    500,
    480,
    NOW()
) ON CONFLICT (id) DO NOTHING;
