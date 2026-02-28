-- Add role column to users table
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin', 'superadmin'));

-- Create an index to quickly find admins
CREATE INDEX idx_users_role ON users(role) WHERE role != 'user';
