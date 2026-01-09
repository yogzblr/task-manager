-- Automation Control Plane Database Schema
-- Multi-tenant, project-aware system of record

-- D1. Tenants
CREATE TABLE IF NOT EXISTS tenants (
  tenant_id VARCHAR(64) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- D2. Projects
CREATE TABLE IF NOT EXISTS projects (
  project_id VARCHAR(64) PRIMARY KEY,
  tenant_id VARCHAR(64) NOT NULL,
  name VARCHAR(255) NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_project_name (tenant_id, name),
  FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
  INDEX idx_projects_tenant (tenant_id)
);

-- D3. Agents
CREATE TABLE IF NOT EXISTS agents (
  agent_id VARCHAR(64) NOT NULL,
  tenant_id VARCHAR(64) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  os VARCHAR(32),
  labels JSON,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (tenant_id, agent_id),
  INDEX idx_agents_project (tenant_id, project_id),
  INDEX idx_agents_tenant (tenant_id),
  FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
  FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE
);

-- D4. Jobs & Leases
CREATE TABLE IF NOT EXISTS jobs (
  job_id CHAR(36) NOT NULL,
  tenant_id VARCHAR(64) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  state ENUM('pending','leased','completed','failed') NOT NULL DEFAULT 'pending',
  lease_owner VARCHAR(64),
  lease_expires_at DATETIME,
  payload JSON NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  completed_at DATETIME,
  PRIMARY KEY (tenant_id, job_id),
  INDEX idx_jobs_project_state (tenant_id, project_id, state),
  INDEX idx_jobs_lease (state, lease_expires_at),
  INDEX idx_jobs_tenant (tenant_id),
  FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
  FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE
);

-- D5. Artifacts
CREATE TABLE IF NOT EXISTS artifacts (
  artifact_id CHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(64) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  sha256 CHAR(64) NOT NULL,
  signature TEXT,
  key_id VARCHAR(32),
  url TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_artifacts_project (tenant_id, project_id),
  INDEX idx_artifacts_tenant (tenant_id),
  INDEX idx_artifacts_sha256 (sha256),
  FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
  FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE
);

-- D6. Users
CREATE TABLE IF NOT EXISTS users (
  user_id VARCHAR(64) PRIMARY KEY,
  tenant_id VARCHAR(64) NOT NULL,
  email VARCHAR(255) NOT NULL,
  display_name VARCHAR(255),
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_user_email (tenant_id, email),
  INDEX idx_users_tenant (tenant_id),
  FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id) ON DELETE CASCADE
);

-- D7. User ↔ Project ↔ Role (RBAC Join)
CREATE TABLE IF NOT EXISTS user_project_roles (
  tenant_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  role ENUM('admin','operator','viewer') NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (tenant_id, user_id, project_id),
  INDEX idx_upr_project (tenant_id, project_id),
  INDEX idx_upr_user (tenant_id, user_id),
  FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
  FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE
);

-- D8. Audit Logs
CREATE TABLE IF NOT EXISTS audit_logs (
  audit_id CHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(64) NOT NULL,
  project_id VARCHAR(64),
  actor_type ENUM('user','agent') NOT NULL,
  actor_id VARCHAR(64) NOT NULL,
  action VARCHAR(128) NOT NULL,
  resource_type VARCHAR(64),
  resource_id VARCHAR(64),
  metadata JSON,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_audit_tenant_time (tenant_id, created_at),
  INDEX idx_audit_project_time (tenant_id, project_id, created_at),
  INDEX idx_audit_actor (tenant_id, actor_type, actor_id),
  FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
  FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE
);
