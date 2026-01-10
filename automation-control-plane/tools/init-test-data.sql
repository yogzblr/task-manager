-- Insert test tenant
INSERT INTO tenants (tenant_id, name, created_at, updated_at) 
VALUES ('test-tenant', 'Test Tenant', NOW(), NOW()) 
ON DUPLICATE KEY UPDATE name='Test Tenant';

-- Insert test project
INSERT INTO projects (project_id, tenant_id, name, created_at, updated_at) 
VALUES ('test-project', 'test-tenant', 'Test Project', NOW(), NOW()) 
ON DUPLICATE KEY UPDATE name='Test Project';
