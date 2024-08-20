CREATE TABLE IF NOT EXISTS key_pairs (
                                         id SERIAL PRIMARY KEY,
                                         private_key BYTEA NOT NULL,
                                         public_key BYTEA NOT NULL,
                                         created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);