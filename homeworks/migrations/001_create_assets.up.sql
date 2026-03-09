-- Migration: Create assets table
-- Version: 001
-- Description: Initial schema for asset management system

CREATE TABLE IF NOT EXISTS assets (
    -- Primary key: UUID for globally unique identifiers
    id UUID PRIMARY KEY,
    
    -- Asset identification
    name VARCHAR(255) NOT NULL,
    
    -- Asset classification
    type VARCHAR(50) NOT NULL,
        
    -- Asset status
    -- CHECK constraint ensures only valid statuses
    status VARCHAR(50) NOT NULL 
        CHECK (status IN ('active', 'inactive')),
    
    -- Audit timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance optimization
-- These speed up common query patterns

-- Index for filtering by type
CREATE INDEX idx_assets_type ON assets(type);

-- Index for filtering by status
CREATE INDEX idx_assets_status ON assets(status);

-- Index for searching by name
-- Supports LIKE/ILIKE queries
CREATE INDEX idx_assets_name ON assets(name);

-- Index for sorting by creation date (DESC order)
-- Most recent assets first
CREATE INDEX idx_assets_created_at ON assets(created_at DESC);

-- Optional: Composite index for common filter combinations
-- Uncomment if you frequently filter by type AND status together
-- CREATE INDEX idx_assets_type_status ON assets(type, status);

/*
🎓 TEACHING NOTES:

1. UUID vs Auto-increment:
   - UUID: globally unique, no coordination needed
   - Auto-increment: simpler but issues in distributed systems
   - We use UUID because it's generated in application layer

2. VARCHAR(255):
   - Common length for names, domains, IPs
   - PostgreSQL: VARCHAR vs TEXT not much difference
   - VARCHAR(n) adds constraint, good for validation

3. CHECK Constraints:
   - Database-level validation
   - Backup for application validation
   - Prevents invalid data even if app has bugs
   - Better than just VARCHAR(50)

4. Timestamps:
   - created_at: set once, never changed
   - updated_at: changed on every update
   - TIMESTAMP includes date + time
   - DEFAULT CURRENT_TIMESTAMP for automatic values

5. Indexes:
   - Speed up SELECT queries
   - Slow down INSERT/UPDATE (tradeoff)
   - Choose based on query patterns
   - Check with EXPLAIN ANALYZE

6. Index Types:
   - B-tree (default): good for equality and range queries
   - GIN: for full-text search (advanced)
   - BRIN: for very large tables (advanced)

7. When to Index:
   - ✅ Foreign keys
   - ✅ Columns in WHERE clause
   - ✅ Columns in ORDER BY
   - ✅ Columns in JOIN conditions
   - ❌ Small tables (< 1000 rows)
   - ❌ Columns rarely queried

8. Migration Best Practices:
   - Always have UP and DOWN scripts
   - Test on dev environment first
   - Can rollback if issues
   - Version numbering: 001, 002, 003...

DEMO QUERIES:

-- Insert sample data
INSERT INTO assets (id, name, type, status, created_at, updated_at)
VALUES 
    (gen_random_uuid(), 'example.com', 'domain', 'active', NOW(), NOW()),
    (gen_random_uuid(), '192.168.1.1', 'ip', 'active', NOW(), NOW());

-- Query with filters (uses indexes!)
SELECT * FROM assets WHERE type = 'domain' AND status = 'active';

-- Search by name (uses idx_assets_name)
SELECT * FROM assets WHERE name ILIKE '%example%';

-- Order by created_at (uses idx_assets_created_at)
SELECT * FROM assets ORDER BY created_at DESC LIMIT 10;
*/
