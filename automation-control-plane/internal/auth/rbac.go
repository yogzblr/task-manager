package auth

import (
	"context"
	"fmt"
)

// Permission represents a fine-grained permission
type Permission string

const (
	PermissionJobRun      Permission = "job:run"
	PermissionJobRead     Permission = "job:read"
	PermissionJobCancel   Permission = "job:cancel"
	PermissionAgentRead   Permission = "agent:read"
	PermissionAgentUpgrade Permission = "agent:upgrade"
	PermissionProjectAdmin Permission = "project:admin"
	PermissionArtifactWrite Permission = "artifact:write"
	PermissionAuditRead    Permission = "audit:read"
)

// Role represents a role with associated permissions
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
)

// RolePermissionMap maps roles to their permissions
var RolePermissionMap = map[Role][]Permission{
	RoleAdmin: {
		PermissionJobRun,
		PermissionJobRead,
		PermissionJobCancel,
		PermissionAgentRead,
		PermissionAgentUpgrade,
		PermissionProjectAdmin,
		PermissionArtifactWrite,
		PermissionAuditRead,
	},
	RoleOperator: {
		PermissionJobRun,
		PermissionJobRead,
		PermissionAgentRead,
	},
	RoleViewer: {
		PermissionJobRead,
		PermissionAgentRead,
	},
}

// RBACAuthorizer authorizes requests based on RBAC
type RBACAuthorizer struct {
	// ProjectRolesGetter retrieves user roles for projects
	ProjectRolesGetter func(ctx context.Context, tenantID, userID string) (map[string][]Role, error)
}

// NewRBACAuthorizer creates a new RBAC authorizer
func NewRBACAuthorizer(projectRolesGetter func(ctx context.Context, tenantID, userID string) (map[string][]Role, error)) *RBACAuthorizer {
	return &RBACAuthorizer{
		ProjectRolesGetter: projectRolesGetter,
	}
}

// Authorize checks if a user has the required permission for a project
func (a *RBACAuthorizer) Authorize(ctx context.Context, claims *JWTClaims, projectID string, requiredPermission Permission) error {
	// Agents are bound to tenant_id + project_id at registration
	if claims.AgentID != "" {
		if claims.TenantID == "" || claims.ProjectID == "" {
			return fmt.Errorf("agent claims must include tenant_id and project_id")
		}
		// Agents can only access their own project
		if claims.ProjectID != projectID {
			return fmt.Errorf("agent not authorized for project %s", projectID)
		}
		// Agents have implicit permissions for their own project
		return nil
	}

	// Users require explicit role-based authorization
	if claims.UserID == "" {
		return fmt.Errorf("user_id required for authorization")
	}

	projectRoles, err := a.ProjectRolesGetter(ctx, claims.TenantID, claims.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user roles: %w", err)
	}

	roles, ok := projectRoles[projectID]
	if !ok || len(roles) == 0 {
		return fmt.Errorf("user not authorized for project %s", projectID)
	}

	// Check if any role has the required permission
	for _, role := range roles {
		permissions, ok := RolePermissionMap[role]
		if !ok {
			continue
		}
		for _, perm := range permissions {
			if perm == requiredPermission {
				return nil
			}
		}
	}

	return fmt.Errorf("permission %s denied for project %s", requiredPermission, projectID)
}

// GetAuthorizedProjects returns the list of projects a user is authorized to access
func (a *RBACAuthorizer) GetAuthorizedProjects(ctx context.Context, claims *JWTClaims) ([]string, error) {
	if claims.AgentID != "" {
		// Agents are bound to a single project
		if claims.ProjectID == "" {
			return nil, fmt.Errorf("agent claims must include project_id")
		}
		return []string{claims.ProjectID}, nil
	}

	if claims.UserID == "" {
		return nil, fmt.Errorf("user_id required")
	}

	projectRoles, err := a.ProjectRolesGetter(ctx, claims.TenantID, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	projects := make([]string, 0, len(projectRoles))
	for projectID := range projectRoles {
		projects = append(projects, projectID)
	}

	return projects, nil
}

// HasPermission checks if a role has a specific permission
func HasPermission(role Role, permission Permission) bool {
	permissions, ok := RolePermissionMap[role]
	if !ok {
		return false
	}
	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// ParsePermission parses a permission string
func ParsePermission(s string) (Permission, error) {
	perm := Permission(s)
	for _, validPerm := range []Permission{
		PermissionJobRun,
		PermissionJobRead,
		PermissionJobCancel,
		PermissionAgentRead,
		PermissionAgentUpgrade,
		PermissionProjectAdmin,
		PermissionArtifactWrite,
		PermissionAuditRead,
	} {
		if perm == validPerm {
			return perm, nil
		}
	}
	return "", fmt.Errorf("invalid permission: %s", s)
}
