package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	AgentID   string `json:"agent_id"`
	TenantID  string `json:"tenant_id"`
	ProjectID string `json:"project_id"`
	jwt.RegisteredClaims
}

func main() {
	// Same secret as in docker-compose
	secret := "change-me-in-production"
	
	// Create claims for the agent
	claims := &Claims{
		AgentID:   "agent-linux-01",
		TenantID:  "test-tenant",
		ProjectID: "test-project",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(365 * 24 * time.Hour)), // 1 year
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	
	// Generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	
	fmt.Println(tokenString)
}
