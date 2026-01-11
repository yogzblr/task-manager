# JWT Token Fix Summary - 2026-01-10

## Problem Identified

The agent couldn't subscribe to Centrifugo channels due to missing JWT claims required for proper authentication and authorization.

## Root Cause Analysis

### Issue 1: Missing `"sub"` Claim ❌ → ✅ FIXED
**Symptom**: Centrifugo logs showed `"user":""`  
**Cause**: JWT token didn't include the `"sub"` (subject) claim required by Centrifugo for user identification.  
**Reference**: [Centrifuge-go client example](https://github.com/centrifugal/centrifuge-go/blob/master/examples/chat/main.go) shows `claims := jwt.MapClaims{"sub": user}`

### Issue 2: Permission Denied for Subscription ❌ → 🔄 IN PROGRESS
**Symptom**: Centrifugo logs showed `"attempt to subscribe without sufficient permission"` even with `"user":"agent-linux-01"`  
**Cause**: Client-side subscriptions require explicit permission or subscription tokens. Server-side subscriptions (via `"channels"` claim) are the recommended approach.  
**Solution**: Add `"channels"` claim to JWT token for automatic server-side subscription.

## Changes Made

### 1. Updated JWT Token Generation Script ✅
**File**: `demo/automation-control-plane/tools/gen-token.py`

**Changes**:
```python
payload = {
    "sub": agent_id,  # Required by Centrifugo for user identification
    "agent_id": agent_id,
    "tenant_id": tenant_id,
    "project_id": project_id,
    "channels": [f"agents.{tenant_id}.{agent_id}"],  # Server-side subscription
    "exp": int(exp.timestamp()),
    "iat": int(now.timestamp())
}
```

**New JWT Token**:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhZ2VudC1saW51eC0wMSIsImFnZW50X2lkIjoiYWdlbnQtbGludXgtMDEiLCJ0ZW5hbnRfaWQiOiJ0ZXN0LXRlbmFudCIsInByb2plY3RfaWQiOiJ0ZXN0LXByb2plY3QiLCJjaGFubmVscyI6WyJhZ2VudHMudGVzdC10ZW5hbnQuYWdlbnQtbGludXgtMDEiXSwiZXhwIjoxNzk5NTYzNTEzLCJpYXQiOjE3NjgwMjc1MTN9.CUmqs_obtkz1xdLUYfdGeHaT4XVNEyJVaEjqKwaIwLk
```

**Decoded Payload**:
```json
{
    "sub": "agent-linux-01",
    "agent_id": "agent-linux-01",
    "tenant_id": "test-tenant",
    "project_id": "test-project",
    "channels": [
        "agents.test-tenant.agent-linux-01"
    ],
    "exp": 1799563513,
    "iat": 1768027513
}
```

### 2. Updated Centrifugo Client (Agent) ✅
**File**: `demo/automation-agent/internal/centrifugo/client.go`

**Changes**: Modified `Subscribe()` method to listen for server-side subscription publications instead of creating client-side subscriptions:

```go
// Subscribe subscribes to the agent's channel
// This method supports server-side subscriptions (via channels claim in JWT)
func (c *Client) Subscribe(handler func([]byte)) error {
	// Listen for server-side subscription publications
	// When channels claim is present in JWT, Centrifugo sets up server-side subscriptions
	c.client.OnPublication(func(e centrifuge.ServerPublicationEvent) {
		handler(e.Data)
	})
	
	return nil
}
```

### 3. Updated Docker Compose Configuration ✅
**File**: `demo/automation-control-plane/deploy/docker-compose.yml`

**Changes**: Updated `agent-linux` service to use the new JWT token with `"sub"` and `"channels"` claims.

### 4. Simplified Centrifugo Configuration ✅
**File**: `demo/automation-control-plane/deploy/centrifugo.json`

**Final Configuration**:
```json
{
  "client": {
    "token": {
      "hmac_secret_key": "change-me-in-production"
    },
    "allowed_origins": ["*"]
  },
  "http_api": {
    "key": "change-me-in-production"
  }
}
```

**Verification**: Centrifugo now shows `"enabled JWT verifiers": "HS256, HS384, HS512"` ✅

## Verification Steps

### Progress So Far ✅
1. ✅ Centrifugo correctly validates JWT with HS256
2. ✅ JWT token includes `"sub"` claim - confirmed by logs showing `"user":"agent-linux-01"`
3. ✅ JWT token includes `"channels"` claim - confirmed by decoding the token
4. ✅ Agent code updated to listen for server-side subscriptions

### Next Steps 🔄
1. Rebuild agent container with updated Centrifugo client code
2. Restart agent with new JWT token containing both `"sub"` and `"channels"` claims
3. Verify agent connects and automatically subscribes via server-side subscription
4. Submit a test job to verify end-to-end workflow

## Expected Behavior

With server-side subscriptions (via `"channels"` claim in JWT):
1. Agent connects to Centrifugo with JWT token
2. Centrifugo automatically subscribes the client to channels listed in `"channels"` claim
3. Agent receives publications via `OnPublication` callback (no manual subscription needed)
4. Control plane can publish job notifications to `agents.test-tenant.agent-linux-01` channel
5. Agent processes the job and workflow completes successfully

## Commands to Complete Fix

```bash
# Rebuild and restart agent with new code and JWT token
cd /mnt/c/Users/yoges/OneDrive/Documents/'My Code'/'Task Manager'/demo/automation-control-plane/deploy
docker compose up -d --build agent-linux

# Wait for agent to start
sleep 10

# Check Centrifugo logs for successful subscription
docker compose logs centrifugo --tail=20 | grep -E "(subscribe|user|agent-linux-01)"

# Check agent logs
docker compose logs agent-linux --tail=20

# Submit test job
cd /mnt/c/Users/yoges/OneDrive/Documents/'My Code'/'Task Manager'
python3 test-linux-workflow.py

# Verify job status
docker compose -f demo/automation-control-plane/deploy/docker-compose.yml exec mysql mysql -u automation -pautomation automation -e "SELECT id, status, created_at FROM jobs ORDER BY created_at DESC LIMIT 1;"
```

## References

- [Centrifuge-go Client Example](https://github.com/centrifugal/centrifuge-go/blob/master/examples/chat/main.go)
- [Centrifugo v6 JWT Token Auth](https://centrifugal.dev/docs/server/authentication)
- [Centrifugo v6 Channel JWT Authorization](https://centrifugal.dev/docs/server/channel_token_auth)
- [Centrifugo v6 Server-side Subscriptions](https://centrifugal.dev/docs/server/server_subs)

## Status

**Current State**: Agent code updated, JWT token generated with all required claims. Build was in progress but canceled.

**Next Action**: Complete the agent rebuild and verify the fix works end-to-end.
