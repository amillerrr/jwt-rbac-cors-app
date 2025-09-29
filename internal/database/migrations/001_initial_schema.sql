-- Migration: 001_initial_schema.sql
-- Description: Initial database schema for authentication system
-- Created: 2025-09-27

-- Drop tables if they exist to start from a clean state
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Create roles table for RBAC (Role-Based Access Control)
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Enhanced users table with authentication fields
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL, -- Store bcrypt hash, never plain text
    email_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Junction table for many-to-many relationship between users and roles
CREATE TABLE user_roles (
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

-- Session management table (if using session-based auth instead of JWT)
CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_accessed TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Enhanced products table with better structure
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX idx_user_sessions_expires ON user_sessions(expires_at);
CREATE INDEX idx_products_user_id ON products(user_id);
CREATE INDEX idx_products_active ON products(is_active);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- Insert default roles
INSERT INTO roles (name, description) VALUES
('admin', 'Full system access'),
('user', 'Standard user access'),
('moderator', 'Content moderation access');

-- Insert sample users (passwords are bcrypt hashes for "password123")
-- Note: These are for development only - NEVER use in production!
INSERT INTO users (name, email, password_hash, email_verified) VALUES
('Alice Admin', 'alice@example.com', '$2a$10$rwu1NKndpjbhnnqqFddBMOtLO8OV0nBWPFYajrckudr./jYL7YxJ6', TRUE),
('Bob User', 'bob@example.com', '$2a$10$rwu1NKndpjbhnnqqFddBMOtLO8OV0nBWPFYajrckudr./jYL7YxJ6', TRUE),
('Charlie Moderator', 'charlie@example.com', '$2a$10$rwu1NKndpjbhnnqqFddBMOtLO8OV0nBWPFYajrckudr./jYL7YxJ6', TRUE);

-- Assign roles to users
INSERT INTO user_roles (user_id, role_id) VALUES
(1, 1), -- Alice is admin
(1, 2), -- Alice is also user
(2, 2), -- Bob is user
(3, 3), -- Charlie is moderator
(3, 2); -- Charlie is also user

-- Insert sample products
INSERT INTO products (name, description, price, user_id) VALUES
('Laptop', 'High-performance laptop for development work', 1200.50, 1),
('Wireless Mouse', 'Ergonomic wireless optical mouse with precision tracking', 25.00, 2),
('Mechanical Keyboard', 'RGB mechanical keyboard with Cherry MX switches', 75.99, 1),
('Monitor', '27-inch 4K monitor with USB-C connectivity', 299.99, 3),
('Webcam', 'HD webcam with auto-focus and noise cancellation', 89.95, 2);

-- Create a function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for auto-updating timestamps
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE users IS 'User accounts with authentication information';
COMMENT ON TABLE roles IS 'System roles for role-based access control';
COMMENT ON TABLE user_roles IS 'Many-to-many mapping between users and roles';
COMMENT ON TABLE products IS 'Product catalog with user ownership';
COMMENT ON TABLE user_sessions IS 'Active user sessions for session-based auth';

COMMENT ON COLUMN users.password_hash IS 'bcrypt hash of user password - never store plaintext';
COMMENT ON COLUMN users.email_verified IS 'Whether the user has verified their email address';
COMMENT ON COLUMN users.is_active IS 'Whether the user account is active - used for soft deletion';
COMMENT ON COLUMN products.price IS 'Product price in decimal format with 2 decimal places';

-- Migration completed successfully
SELECT 'Migration 001_initial_schema.sql completed successfully' as result;
