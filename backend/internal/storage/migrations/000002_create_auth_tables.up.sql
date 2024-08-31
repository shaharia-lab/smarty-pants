-- Create users table
CREATE TABLE IF NOT EXISTS users (
    uuid UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL,
    roles TEXT[] DEFAULT ARRAY['user'],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS key_pairs (
    id SERIAL PRIMARY KEY,
    private_key BYTEA NOT NULL,
    public_key BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users (uuid, name, email, status, roles, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000000',
    'Anonymous User',
    'user@example.com',
    'active',
    ARRAY['admin'],
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);