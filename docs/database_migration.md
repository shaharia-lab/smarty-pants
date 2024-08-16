# Database Migration System

## Overview

The database migration system is a crucial component of our application, ensuring that the database schema remains consistent and up-to-date across different environments and versions of the application. It provides a structured way to evolve the database schema over time, allowing for easy updates and rollbacks when necessary.

## Key Components

### 1. Migration Manager

The `Manager` struct is the core of the migration system. It orchestrates the entire migration process, including:

- Ensuring the migration table exists
- Acquiring and releasing migration locks
- Running migrations
- Handling migration errors

### 2. Migrator Interface

The `Migrator` interface defines the contract for implementing database-specific migration operations. It includes methods for:

- Migrating
- Rolling back
- Getting the current version
- Ensuring the migration table exists
- Acquiring and releasing migration locks

### 3. Migration Struct

The `Migration` struct represents a single migration and contains:

- Version: A string identifier for the migration
- Up: A function to apply the migration
- Down: A function to revert the migration

### 4. PostgreSQL Migrations

The `postgreSQLMigrations` slice contains all the migrations for PostgreSQL. Each migration is defined with an Up and Down function to apply and revert changes respectively.

## Migration Process

1. **Initialization**: The `NewMigrationManager` function creates a new `Manager` instance with the provided `Migrator` and logger.

2. **Running Migrations**: The `Run` method of the `Manager` executes the migration process:
    - Ensures the migration table exists
    - Acquires a migration lock
    - Runs all pending migrations
    - Releases the migration lock

3. **Locking Mechanism**: The system uses a locking mechanism to prevent concurrent migrations, ensuring data integrity.

4. **Error Handling**: The system handles various types of errors, including:
    - Dirty migrations
    - Failed migrations
    - Lock acquisition failures

## Implemented Migrations

Currently, the system includes one migration (version 0.0.1) that sets up the initial database schema:

- Creates necessary tables (datasources, documents, content_parts, etc.)
- Sets up indexes for efficient querying
- Enables the pgvector extension for vector operations

## FAQs

1. **Q: How are concurrent migrations prevented?**
   A: The system uses a locking mechanism. Before running migrations, it attempts to acquire a lock. If the lock is already held, it assumes another instance is running migrations and exits.

2. **Q: What happens if a migration fails?**
   A: If a migration fails, the system will stop the migration process and return an error. The database will be left in a "dirty" state, which will need to be resolved manually.

3. **Q: Can migrations be rolled back?**
   A: Yes, each migration includes a "Down" function that reverts the changes made by the "Up" function. However, the current implementation doesn't provide an automatic rollback feature.

4. **Q: How are "dirty" migrations handled?**
   A: The system detects dirty migrations (migrations that failed partway through) and reports them as errors. These need to be resolved manually.

5. **Q: What happens if the migration table doesn't exist?**
   A: The system attempts to create the migration table before running any migrations. If this fails, the migration process is aborted.

6. **Q: How are migration versions managed?**
   A: Each migration has a version string. The system keeps track of the last successfully applied migration version in the database.

## Handled Cases

1. **Successful Migration**: All migrations are applied successfully.
2. **Failed Lock Acquisition**: Another instance is already running migrations.
3. **Lock Acquisition Error**: An error occurs while trying to acquire the lock.
4. **Migration Table Creation Failure**: The system fails to create or ensure the existence of the migration table.
5. **Migration Failure**: A specific migration fails to apply.
6. **Lock Release Failure**: The system fails to release the migration lock.

## Best Practices

1. Always test migrations in a non-production environment before applying them to production.
2. Keep migrations small and focused on specific changes to make troubleshooting easier.
3. Ensure that both "Up" and "Down" functions are implemented for each migration.
4. Use meaningful and incrementing version numbers for migrations.
5. Handle errors gracefully and provide clear error messages for easier debugging.

## Future Improvements

1. Implement an automatic rollback feature for failed migrations.
2. Add a dry-run option to preview migration changes without applying them.
3. Implement a mechanism to resolve dirty migrations automatically in simple cases.
4. Add more detailed logging and monitoring for the migration process.
5. Implement a CLI tool for running migrations manually and checking migration status.

By following these guidelines and understanding the migration system, you can ensure smooth database schema evolution throughout your project's lifecycle.