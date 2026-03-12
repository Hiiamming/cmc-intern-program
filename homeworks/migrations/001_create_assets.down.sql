-- Rollback migration: Drop assets table
-- Version: 001
-- Description: Reverts the initial schema creation

-- Drop indexes first (PostgreSQL requires this)
DROP INDEX IF EXISTS idx_assets_created_at;
DROP INDEX IF EXISTS idx_assets_name;
DROP INDEX IF EXISTS idx_assets_status;
DROP INDEX IF EXISTS idx_assets_type;

-- Drop the main table
DROP TABLE IF EXISTS assets;

/*
🎓 TEACHING NOTES:

1. Migration Rollback:
   - DOWN migration = undo UP migration
   - Essential for safe deployments
   - If UP fails, run DOWN to clean up

2. Drop Order:
   - Drop indexes before table
   - PostgreSQL: indexes are separate objects
   - Dropping table also drops indexes, but explicit is better

3. IF EXISTS:
   - Prevents errors if object doesn't exist
   - Makes migrations idempotent (can run multiple times)
   - Good for development environments

4. When to Run DOWN:
   - Migration caused issues
   - Need to modify schema structure
   - Testing migration process
   - Cleaning up dev environment

5. Production Considerations:
   - NEVER run DOWN on production without backup!
   - Data loss: DROP TABLE deletes all data
   - Better: use migration tools (golang-migrate, flyway)
   - Version control: track which migrations ran

MANUAL MIGRATION WORKFLOW:

Development:
1. Write UP migration
2. Run: psql -U postgres -d mini_asm -f 001_create_assets.up.sql
3. Test application
4. If issues: psql -U postgres -d mini_asm -f 001_create_assets.down.sql
5. Fix and repeat

Production:
1. Test migrations on staging first
2. Backup database
3. Run migration
4. Verify application works
5. Keep DOWN script ready (just in case)

MIGRATION TOOLS (Buổi 6):
- golang-migrate: CLI tool for Go projects
- Tracks version in schema_migrations table
- Automatic UP/DOWN execution
- Works with CI/CD pipelines
*/
