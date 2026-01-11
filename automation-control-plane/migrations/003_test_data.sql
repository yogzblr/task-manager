-- Insert test tenant and project for agent registration
INSERT INTO tenants (tenant_id, name) VALUES ('test-tenant', 'Test Tenant')
ON DUPLICATE KEY UPDATE name = 'Test Tenant';

INSERT INTO projects (project_id, tenant_id, name) VALUES ('test-project', 'test-tenant', 'Test Project')
ON DUPLICATE KEY UPDATE name = 'Test Project';
