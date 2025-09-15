-- Fix database issues for testing

-- Create sites table if it doesn't exist
CREATE TABLE IF NOT EXISTS sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    location VARCHAR(500),
    client_name VARCHAR(255),
    project_number VARCHAR(100),
    installation_date TIMESTAMPTZ,
    capacity FLOAT,
    status VARCHAR(50) DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Insert test site with correct columns
INSERT INTO sites (id, site_code, name, address, country, total_capacity_kw, number_of_inverters, installation_date)
VALUES ('123e4567-e89b-12d3-a456-426614174000', 'SFA001', 'Solar Farm Alpha', 'California, USA', 'US', 25500, 10, '2023-01-15')
ON CONFLICT (id) DO NOTHING;

-- Add missing column to site_components if needed
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                  WHERE table_name = 'site_components' 
                  AND column_name = 'sort_order') THEN
        ALTER TABLE site_components ADD COLUMN sort_order INTEGER DEFAULT 0;
    END IF;
END $$;

-- Create user_sessions table if it doesn't exist
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token VARCHAR(500) NOT NULL,
    refresh_token VARCHAR(500),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create users table if it doesn't exist
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Insert a test user (check actual user table schema)
-- Skip if table columns don't match

-- Verify the setup
SELECT 'Database setup completed!' as status;
SELECT COUNT(*) as site_count FROM sites WHERE id = '123e4567-e89b-12d3-a456-426614174000';