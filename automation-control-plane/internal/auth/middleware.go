package auth

import (
	"context"
	"fmt"
	"net/http"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	ContextKeyClaims ContextKey = "claims"
	ContextKeyTenant ContextKey = "tenant_id"
)

// AuthMiddleware validates JWT tokens and adds claims to context
func AuthMiddleware(validator *JWTValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			claims, err := validator.ValidateToken(r.Context(), authHeader)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), ContextKeyClaims, claims)
			ctx = context.WithValue(ctx, ContextKeyTenant, claims.TenantID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission middleware checks if user has required permission
func RequirePermission(authorizer *RBACAuthorizer, projectIDExtractor func(*http.Request) string, permission Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ContextKeyClaims).(*JWTClaims)
			if !ok {
				http.Error(w, "missing claims in context", http.StatusInternalServerError)
				return
			}

			projectID := projectIDExtractor(r)
			if projectID == "" {
				http.Error(w, "project_id required", http.StatusBadRequest)
				return
			}

			if err := authorizer.Authorize(r.Context(), claims, projectID, permission); err != nil {
				http.Error(w, fmt.Sprintf("authorization failed: %v", err), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClaimsFromContext extracts JWT claims from context
func GetClaimsFromContext(ctx context.Context) (*JWTClaims, bool) {
	claims, ok := ctx.Value(ContextKeyClaims).(*JWTClaims)
	return claims, ok
}

// GetTenantIDFromContext extracts tenant ID from context
func GetTenantIDFromContext(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(ContextKeyTenant).(string)
	return tenantID, ok
}
