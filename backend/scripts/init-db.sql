-- Initialize Engramiq Database
-- This script runs when the PostgreSQL container first starts

-- Install required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector;

-- Create additional schemas if needed
-- CREATE SCHEMA IF NOT EXISTS analytics;

-- Set up initial configuration
SET timezone = 'UTC';

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE engramiq TO engramiq;
GRANT ALL ON SCHEMA public TO engramiq;

-- Log successful initialization
SELECT 'Database initialized successfully with pgvector extension' as status;