# Solar Asset Database Schema

This document describes the complete database schema for the Engramiq Solar Asset Reporting Agent, including the migration system and pre-populated Supporting-Documents data.

## Overview

The database is built on **PostgreSQL 15** with **pgvector extension** for semantic search capabilities. The schema supports:

- **Complete Migration System**: Production-ready migrations with tracking and rollback
- **Solar Asset Management**: Sites, components, relationships, and hierarchies  
- **Document Processing**: Field service reports with AI-powered analysis
- **Enhanced Query System**: Natural language processing with source attribution
- **Supporting-Documents Integration**: Pre-loaded site S2367 with 46+ SOLECTRIA inverters

## Database Extensions

```sql
-- Required PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";
```

## Migration System Tables

### migration_records
Tracks applied database migrations to ensure consistency and prevent duplicates.

```sql
CREATE TABLE migration_records (
    id VARCHAR PRIMARY KEY,              -- Migration ID (e.g., "20240915000001")
    name VARCHAR NOT NULL,               -- Human-readable migration name  
    applied_at TIMESTAMP NOT NULL       -- When migration was applied
);
```

**Populated Data**:
- `20240915000001`: "Populate site S2367 with inverter data" - Applied automatically on startup

## Core Domain Tables

### sites
Solar installation sites with capacity and location information.

```sql
CREATE TABLE sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_code VARCHAR(50) UNIQUE NOT NULL,      -- Human-readable site identifier
    name VARCHAR(255) NOT NULL,
    address TEXT,
    country VARCHAR(2) DEFAULT 'US',
    total_capacity_kw DECIMAL(10,2),
    number_of_inverters INTEGER,
    installation_date DATE,
    site_metadata JSONB DEFAULT '{}',           -- Flexible metadata storage
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP                        -- Soft delete support
);

-- Indexes
CREATE INDEX idx_sites_site_code ON sites(site_code);
CREATE INDEX idx_sites_country ON sites(country);
CREATE UNIQUE INDEX uni_sites_site_code ON sites(site_code);
```

**Pre-populated Data**:
- **Site S2367**: "Combined Site - Educational Campus" with 46+ SOLECTRIA inverters from Supporting-Documents

### site_components  
Equipment components within solar installations (inverters, combiners, panels, etc.).

```sql
-- Custom ENUM types
CREATE TYPE component_type AS ENUM (
    'inverter', 'combiner', 'panel', 'transformer', 
    'meter', 'switchgear', 'monitoring', 'other'
);

CREATE TABLE site_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    external_id VARCHAR(255),                   -- External system identifier
    component_type component_type NOT NULL,
    name VARCHAR(255) NOT NULL,                 -- Component name/identifier
    label VARCHAR(255),                         -- Display label
    level INTEGER DEFAULT 0,                    -- Hierarchy level
    group_name VARCHAR(255),                    -- Grouping identifier
    
    -- Technical specifications
    specifications JSONB DEFAULT '{}',          -- Manufacturer, model, capacity, etc.
    electrical_data JSONB DEFAULT '{}',         -- Voltage, current, power specifications
    physical_data JSONB DEFAULT '{}',           -- Location, coordinates, dimensions
    
    -- Documentation
    drawing_title VARCHAR(500),
    drawing_number VARCHAR(100), 
    revision VARCHAR(50),
    revision_date DATE,
    
    -- Location and relationships
    spatial_id UUID,                            -- Reference to spatial data
    coordinates VARCHAR(100),                   -- Simple coordinate storage
    embedding VECTOR(1536),                     -- AI embedding for semantic search
    
    -- Status tracking
    current_status VARCHAR(50) DEFAULT 'operational',
    last_maintenance_date DATE,
    next_maintenance_date DATE,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Indexes
CREATE INDEX idx_components_site_type ON site_components(site_id, component_type);
CREATE INDEX idx_components_name ON site_components(name);
CREATE INDEX idx_components_status ON site_components(current_status);
CREATE INDEX idx_components_embedding ON site_components USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_components_specifications ON site_components USING gin(specifications);
CREATE INDEX idx_components_electrical_data ON site_components USING gin(electrical_data);
```

**Pre-populated Data** (from inverter_nodes.json):
- **46+ SOLECTRIA Inverters**: Ground mount (PVI 75TL, 75kW each) and rooftop (S5-GR3P15K, 15kW each)
- **Complete Specifications**: Manufacturer, model, capacity, serial numbers
- **Electrical Data**: Max DC/AC power, efficiency, MPPT channels, voltage ranges
- **Physical Location**: Area (Ground/Rooftop), row/position coordinates, spatial IDs

### component_relationships
Relationships between components (power flow, control, monitoring, etc.).

```sql
CREATE TYPE relationship_type AS ENUM (
    'connects_to', 'powers', 'controls', 'monitors',
    'parent_child', 'same_string', 'same_combiner'
);

CREATE TABLE component_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_component_id UUID NOT NULL REFERENCES site_components(id),
    child_component_id UUID NOT NULL REFERENCES site_components(id),
    relationship_type relationship_type NOT NULL,
    relationship_data JSONB DEFAULT '{}',       -- Additional relationship metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_relationships_parent ON component_relationships(parent_component_id);
CREATE INDEX idx_relationships_child ON component_relationships(child_component_id);
CREATE INDEX idx_relationships_type ON component_relationships(relationship_type);
```

## Document Management Tables

### documents
Uploaded documents (field service reports, emails, manuals, etc.) with AI processing.

```sql
CREATE TYPE document_type AS ENUM (
    'field_service_report', 'email', 'meeting_transcript',
    'work_order', 'inspection_report', 'warranty_claim',
    'contract', 'manual', 'drawing', 'other'
);

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    document_type document_type NOT NULL,
    file_path VARCHAR(1000),                    -- Storage path
    file_size BIGINT,
    file_hash VARCHAR(64),                      -- Content deduplication
    mime_type VARCHAR(100),
    
    -- Content processing
    raw_content TEXT,                           -- Extracted text content
    processed_content TEXT,                     -- Cleaned/processed content
    embedding VECTOR(1536),                     -- AI embedding for semantic search
    processing_status VARCHAR(50) DEFAULT 'pending', -- pending/processing/completed/failed
    processing_error TEXT,
    
    -- Search optimization
    content_vector TSVECTOR,                    -- Full-text search vector
    
    -- Metadata
    document_metadata JSONB DEFAULT '{}',       -- Flexible metadata
    upload_user_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Indexes
CREATE INDEX idx_documents_site ON documents(site_id);
CREATE INDEX idx_documents_type ON documents(document_type);
CREATE INDEX idx_documents_status ON documents(processing_status);
CREATE INDEX idx_documents_hash ON documents(file_hash);
CREATE INDEX idx_documents_fts ON documents USING gin(content_vector);
CREATE INDEX idx_documents_embedding ON documents USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Trigger for automatic content_vector updates
CREATE OR REPLACE FUNCTION update_content_vector() RETURNS trigger AS $$
BEGIN
    NEW.content_vector := to_tsvector('english', COALESCE(NEW.title, '') || ' ' || COALESCE(NEW.processed_content, ''));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER documents_content_vector_trigger
BEFORE INSERT OR UPDATE OF title, processed_content ON documents
FOR EACH ROW
EXECUTE FUNCTION update_content_vector();
```

**Pre-populated Data** (from Sample-Reports.pdf):
- **Field Service Reports**: Arc protection issues, inverter replacements, wire management
- **Automatic Processing**: All documents processed with embeddings for immediate querying
- **Searchable Content**: Ready for enhanced natural language queries

## AI-Powered Analysis Tables

### extracted_actions
Maintenance actions extracted from documents using AI analysis.

```sql
CREATE TYPE action_type AS ENUM (
    'maintenance', 'replacement', 'troubleshoot', 'inspection',
    'repair', 'testing', 'installation', 'commissioning',
    'fault_clearing', 'monitoring', 'cleaning', 'other'
);

CREATE TYPE action_status AS ENUM (
    'planned', 'in_progress', 'completed', 'cancelled',
    'on_hold', 'requires_follow_up'
);

CREATE TABLE extracted_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    site_id UUID NOT NULL REFERENCES sites(id),
    
    -- Action details
    action_type action_type NOT NULL,
    description TEXT NOT NULL,
    action_date DATE,
    status action_status DEFAULT 'completed',
    
    -- Component associations
    component_type VARCHAR(50),
    component_names TEXT[],                     -- Array of component identifiers
    
    -- Work details  
    technician_names TEXT[],                    -- Array of technician names
    work_order_number VARCHAR(100),
    duration_hours DECIMAL(5,2),
    
    -- Technical data
    measurements JSONB DEFAULT '{}',            -- Voltage, current, resistance measurements
    parts_used JSONB DEFAULT '{}',              -- Parts and materials
    
    -- AI analysis metadata
    confidence_score DECIMAL(3,2),              -- AI extraction confidence (0.0-1.0)
    embedding VECTOR(1536),                     -- Semantic embedding
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_actions_document ON extracted_actions(document_id);
CREATE INDEX idx_actions_site_date ON extracted_actions(site_id, action_date);
CREATE INDEX idx_actions_type ON extracted_actions(action_type);
CREATE INDEX idx_actions_status ON extracted_actions(status);
CREATE INDEX idx_actions_technicians ON extracted_actions USING gin(technician_names);
CREATE INDEX idx_actions_measurements ON extracted_actions USING gin(measurements);
CREATE INDEX idx_actions_embedding ON extracted_actions USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
```

### action_components
Links between extracted actions and specific components.

```sql
CREATE TABLE action_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action_id UUID NOT NULL REFERENCES extracted_actions(id) ON DELETE CASCADE,
    component_id UUID REFERENCES site_components(id),
    component_name VARCHAR(255),                -- For components not in our database
    is_primary BOOLEAN DEFAULT FALSE,           -- Primary component for this action
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_action_components_action ON action_components(action_id);
CREATE INDEX idx_action_components_component ON action_components(component_id);
```

## Enhanced Query System Tables

### user_queries
Natural language queries with AI-powered analysis and source attribution.

```sql
CREATE TABLE user_queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id),
    user_id UUID,                               -- Future user management
    
    -- Query details
    query_text TEXT NOT NULL,
    query_type VARCHAR(100),                    -- maintenance_history, component_status, etc.
    enhanced BOOLEAN DEFAULT FALSE,             -- Use enhanced PRD features
    
    -- AI processing results
    answer TEXT,
    confidence_score DECIMAL(3,2),              -- Response confidence (0.0-1.0) 
    extracted_entities JSONB DEFAULT '{}',      -- Components, dates, technicians, etc.
    related_concepts TEXT[],                    -- Related search terms
    
    -- Response metadata
    response_type VARCHAR(50),                  -- summary, maintenance_history, error, etc.
    no_hallucination BOOLEAN DEFAULT TRUE,      -- Source-based response validation
    processing_time_ms INTEGER,                 -- Performance tracking
    
    -- Search and analysis
    embedding VECTOR(1536),                     -- Query embedding for similarity
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_queries_site ON user_queries(site_id);
CREATE INDEX idx_queries_user ON user_queries(user_id);
CREATE INDEX idx_queries_type ON user_queries(query_type);
CREATE INDEX idx_queries_enhanced ON user_queries(enhanced);
CREATE INDEX idx_queries_embedding ON user_queries USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_queries_entities ON user_queries USING gin(extracted_entities);
```

### query_sources
Source attribution linking queries to supporting documents.

```sql
CREATE TABLE query_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_id UUID NOT NULL REFERENCES user_queries(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id),
    
    -- Source details
    relevant_excerpt TEXT,                      -- Specific text excerpt used
    relevance_score DECIMAL(3,2),               -- How relevant this source is (0.0-1.0)
    citation VARCHAR(500),                      -- Formatted citation
    page_number INTEGER,                        -- Page reference (if applicable)
    section_title VARCHAR(255),                 -- Document section
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes  
CREATE INDEX idx_query_sources_query ON query_sources(query_id);
CREATE INDEX idx_query_sources_document ON query_sources(document_id);
CREATE INDEX idx_query_sources_relevance ON query_sources(relevance_score);
```

## Timeline and Event Tracking

### site_events
Timeline of events and activities across solar installations.

```sql
CREATE TYPE event_type AS ENUM (
    'maintenance_scheduled', 'maintenance_completed',
    'fault_occurred', 'fault_cleared',
    'replacement_scheduled', 'replacement_completed',
    'inspection_scheduled', 'inspection_completed', 
    'warranty_claim', 'performance_alert',
    'contract_milestone', 'other'
);

CREATE TYPE event_priority AS ENUM ('low', 'medium', 'high', 'critical');

CREATE TABLE site_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    
    -- Event details
    event_type event_type NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority event_priority DEFAULT 'medium',
    
    -- Timing
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_minutes INTEGER,
    
    -- Associations
    affected_component_ids UUID[],              -- Components involved
    related_document_id UUID REFERENCES documents(id),
    related_action_id UUID REFERENCES extracted_actions(id),
    
    -- Personnel
    assigned_technicians TEXT[],
    responsible_user_id UUID,
    
    -- Metadata
    event_metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_events_site_timeline ON site_events(site_id, start_time, end_time);
CREATE INDEX idx_events_type ON site_events(event_type);
CREATE INDEX idx_events_priority ON site_events(priority);
CREATE INDEX idx_events_affected_components ON site_events USING gin(affected_component_ids);
```

## User Management Tables (Ready for Implementation)

### users
User accounts and authentication.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,       -- bcrypt hash
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) DEFAULT 'viewer',          -- admin, manager, technician, viewer
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Profile
    organization VARCHAR(255),
    phone VARCHAR(50),
    user_metadata JSONB DEFAULT '{}',
    
    -- Authentication
    email_verified BOOLEAN DEFAULT FALSE,
    last_login TIMESTAMP,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Indexes
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_active ON users(is_active);
```

### refresh_tokens  
JWT refresh token management.

```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,          -- Hashed refresh token
    expires_at TIMESTAMP NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE,
    
    -- Security tracking
    issued_ip VARCHAR(45),                      -- IPv4/IPv6 support
    user_agent TEXT,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at);
```

## Analytics and Performance Tables

### query_analytics
Query performance and usage analytics.

```sql
CREATE TABLE query_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id),
    query_id UUID REFERENCES user_queries(id),
    user_id UUID REFERENCES users(id),
    
    -- Performance metrics
    processing_time_ms INTEGER NOT NULL,
    source_retrieval_time_ms INTEGER,
    llm_processing_time_ms INTEGER,
    total_sources_found INTEGER,
    sources_used INTEGER,
    
    -- Quality metrics
    confidence_score DECIMAL(3,2),
    user_satisfaction INTEGER,                  -- 1-5 rating (future feature)
    query_successful BOOLEAN DEFAULT TRUE,
    
    -- Usage patterns
    query_category VARCHAR(100),                -- maintenance, troubleshooting, reporting
    response_length INTEGER,                    -- Character count of response
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_analytics_site_date ON query_analytics(site_id, created_at);
CREATE INDEX idx_analytics_performance ON query_analytics(processing_time_ms);
CREATE INDEX idx_analytics_success ON query_analytics(query_successful);
```

## Database Management

### Migration System Usage

```sql
-- Check applied migrations
SELECT * FROM migration_records ORDER BY applied_at;

-- Verify Supporting-Documents integration
SELECT site_code, name, number_of_inverters 
FROM sites WHERE site_code = 'S2367';

-- Count migrated components
SELECT COUNT(*) as total_components, 
       specifications->>'manufacturer' as manufacturer
FROM site_components 
WHERE site_id = (SELECT id FROM sites WHERE site_code = 'S2367')
GROUP BY specifications->>'manufacturer';
```

### Performance Optimization

**Connection Pooling**:
- Maximum connections: 100
- Idle connections: 10
- Connection timeout: 30 seconds

**Query Optimization**:
- All foreign keys indexed
- Composite indexes for common query patterns
- GIN indexes for JSONB and array columns
- IVFFlat indexes for vector similarity search

**Vector Search Configuration**:
- Lists parameter: 100 (balanced speed/accuracy)
- Distance metric: Cosine similarity
- Index maintenance: Automatic

## Security Considerations

### Access Control
- Row-level security ready for multi-tenant deployment
- Site-based access control through foreign keys
- User role-based permissions structure

### Data Protection
- Sensitive information sanitization in content processing
- Audit trail through created_at/updated_at timestamps
- Soft delete support for data recovery

### Performance Security
- Query timeout limits to prevent resource exhaustion
- Rate limiting through application layer
- Input validation for all text fields

## Supporting-Documents Integration Summary

The database comes pre-loaded with production-ready data:

**Site S2367**: Complete solar installation with:
- 46+ SOLECTRIA inverters with full specifications
- Ground mount (PVI 75TL, 75kW) and rooftop (S5-GR3P15K, 15kW) systems
- Spatial IDs, electrical data, and physical locations

**Field Service Reports**: Processed from Sample-Reports.pdf:
- Inverter 40 arc protection issues
- Inverter 31 replacement procedures  
- Wire management and maintenance activities

**Enhanced Query System**: Ready for immediate use:
- Natural language queries with source attribution
- Professional behavior guards and content filtering
- Complete audit trail and traceability

This schema provides a complete, production-ready foundation for solar asset management with advanced AI capabilities and comprehensive data tracking.