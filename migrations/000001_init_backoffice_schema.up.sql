-- Migration: 000001_init_backoffice_schema.up.sql
-- Description: Create schemas and base tables for backoffice service

CREATE SCHEMA IF NOT EXISTS schema_backoffice;
CREATE SCHEMA IF NOT EXISTS schema_ledger;
CREATE SCHEMA IF NOT EXISTS schema_burning;

-- Enable UUID extension if not enabled
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 1. Admin Users
CREATE TABLE IF NOT EXISTS schema_backoffice.admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    role VARCHAR(30) NOT NULL CHECK (role IN ('SUPER_ADMIN', 'CAMPAIGN_SPEC', 'OPS_CS', 'FINANCE')),
    permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Dynamic Rules (Earning Rule Builder)
CREATE TABLE IF NOT EXISTS schema_backoffice.dynamic_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_code VARCHAR(50) UNIQUE NOT NULL,
    rule_name VARCHAR(100) NOT NULL,
    description TEXT,
    rule_type VARCHAR(30) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    priority INTEGER NOT NULL DEFAULT 0,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    conditions_json JSONB NOT NULL,
    action_json JSONB NOT NULL,
    created_by UUID REFERENCES schema_backoffice.admin_users(id),
    updated_by UUID REFERENCES schema_backoffice.admin_users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Tier Definitions
CREATE TABLE IF NOT EXISTS schema_backoffice.tier_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier_code VARCHAR(30) UNIQUE NOT NULL,
    tier_name VARCHAR(50) NOT NULL,
    min_points INTEGER NOT NULL DEFAULT 0,
    max_points INTEGER,
    multiplier DECIMAL(5,2) NOT NULL DEFAULT 1.0,
    color_hex VARCHAR(7),
    icon_url TEXT,
    perks JSONB NOT NULL DEFAULT '[]'::jsonb,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_by UUID REFERENCES schema_backoffice.admin_users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Pending Actions (Maker-Checker Queue)
CREATE TABLE IF NOT EXISTS schema_backoffice.admin_pending_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action_type VARCHAR(50) NOT NULL,
    target_resource VARCHAR(100) NOT NULL,
    target_id VARCHAR(100),
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    maker_id UUID NOT NULL REFERENCES schema_backoffice.admin_users(id),
    checker_id UUID REFERENCES schema_backoffice.admin_users(id),
    checker_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

-- 5. Audit Logs (Immutable)
CREATE TABLE IF NOT EXISTS schema_backoffice.admin_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    admin_id UUID,
    action VARCHAR(100) NOT NULL,
    target_resource VARCHAR(200),
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. Campaigns
CREATE TABLE IF NOT EXISTS schema_backoffice.campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(30) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    rules JSONB NOT NULL DEFAULT '[]'::jsonb,
    target_tiers JSONB NOT NULL DEFAULT '[]'::jsonb,
    budget_points INTEGER,
    created_by UUID REFERENCES schema_backoffice.admin_users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 7. Partners
CREATE TABLE IF NOT EXISTS schema_backoffice.partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_code VARCHAR(30) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(30) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    contact_email VARCHAR(100),
    contact_phone VARCHAR(20),
    settlement_rate DECIMAL(5,4) NOT NULL DEFAULT 0.01,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_dynamic_rules_status ON schema_backoffice.dynamic_rules(status);
CREATE INDEX IF NOT EXISTS idx_dynamic_rules_type ON schema_backoffice.dynamic_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_pending_actions_status ON schema_backoffice.admin_pending_actions(status);
CREATE INDEX IF NOT EXISTS idx_audit_logs_admin_id ON schema_backoffice.admin_audit_logs(admin_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON schema_backoffice.admin_audit_logs(created_at DESC);
