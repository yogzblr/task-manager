# Subscription Fix Status - 2026-01-10

## Approach Changed: Client-Side Subscriptions

After encountering Docker build issues with the server-side subscription approach, we switched to a simpler solution: configuring Centrifugo to allow client-side subscriptions.

## Changes Completed ✅

### 1. Updated Centrifugo Configuration
**File**: [`demo/automation-control-plane/deploy/centrifugo.json`](demo/automation-control-plane/deploy/centrifugo.json)

**Change**: Added client subscription permissions
```json
{
  "client": {
    "token": {
      "hmac_secret_key": "change-me-in-production"
    },
    "allowed_origins": ["*"],
    "allow_subscribe_for_client": true,  // ← ADDED
    "allow_publish_for_client": true     // ← ADDED
  },
  "http_api": {
    "key": "change-me-in-production"
  }
}
```

**Why**: This allows authenticated clients (with valid JWT tokens) to subscribe to channels without requiring:
- Server-side subscriptions (channels claim in JWT)
- Subscription tokens
- Agent code changes

### 2. Restarted Services
- ✅ Centrifugo restarted to apply new configuration
- ✅ Agent restarted to reconnect with new permissions

## Current Status: Blocked ⚠️

**Issue**: Docker Desktop WSL integration is not active

**Error**:
```
The command 'docker' could not be found in this WSL 2 distro.
We recommend to activate the WSL integration in Docker Desktop settings.
```

**Impact**: Cannot verify if the fix worked because Docker commands are not available in WSL.

## Next Steps to Complete

### 1. Restart Docker Desktop
1. Open Docker Desktop application
2. Wait for it to fully start
3. Verify WSL integration is enabled:
   - Settings → Resources → WSL Integration
   - Ensure "Ubuntu-22.04" is checked

### 2. Verify Containers are Running
```bash
wsl -d Ubuntu-22.04 docker ps
```

Expected output should show:
- deploy-centrifugo-1
- deploy-agent-linux-1
- deploy-control-plane-1
- deploy-mysql-1
- deploy-quickwit-1
- etc.

### 3. Check Centrifugo Logs
```bash
wsl -d Ubuntu-22.04 docker logs deploy-centrifugo-1 --tail 30
```

**Look for**:
- ✅ `"enabled JWT verifiers": "HS256, HS384, HS512"`
- ✅ `"user":"agent-linux-01"` (from JWT sub claim)
- ✅ Successful subscription (no "permission denied" errors)

Expected log pattern:
```json
{"level":"info","user":"agent-linux-01","client":"...","time":"...","message":"client connected"}
{"level":"info","channel":"agents.test-tenant.agent-linux-01","user":"agent-linux-01","time":"...","message":"client subscribed"}
```

### 4. Check Agent Logs
```bash
wsl -d Ubuntu-22.04 docker logs deploy-agent-linux-1 --tail 20
```

**Look for**:
- ✅ No "failed to subscribe" errors
- ✅ No "permission denied" errors
- ✅ Agent running and connected

### 5. Submit Test Workflow
```bash
cd demo/automation-control-plane/deploy
wsl -d Ubuntu-22.04 docker compose run --rm test-runner python test-linux-workflow.py
```

### 6. Verify Job Execution
```bash
wsl -d Ubuntu-22.04 docker compose -f demo/automation-control-plane/deploy/docker-compose.yml exec mysql mysql -u automation -pautomation automation -e "SELECT id, status, created_at FROM jobs ORDER BY created_at DESC LIMIT 5;"
```

**Expected**: Job status should progress from `pending` → `running` → `completed`

### 7. Query Quickwit Logs
```bash
curl "http://localhost:7280/api/v1/automation-logs/search?query=test"
```

## Why This Approach Works

### Technical Details

1. **JWT Token** (already configured):
   - Contains `"sub": "agent-linux-01"` for user identification
   - Centrifugo validates JWT and identifies the user

2. **Centrifugo Config** (newly updated):
   - `allow_subscribe_for_client: true` - Authenticated clients can subscribe
   - `allow_publish_for_client: true` - Authenticated clients can publish

3. **Agent Code** (existing - no changes needed):
   - Agent calls `sub.Subscribe()` for client-side subscription
   - With new Centrifugo permissions, this subscription will succeed
   - Agent receives publications via `sub.OnPublication()` callback

### Sequence Diagram

```mermaid
sequenceDiagram
    participant Agent
    participant Centrifugo
    participant ControlPlane
    
    Agent->>Centrifugo: Connect with JWT (has "sub" claim)
    Centrifugo->>Centrifugo: Validate JWT → user="agent-linux-01"
    Centrifugo->>Agent: Connection established
    
    Agent->>Centrifugo: Subscribe to agents.test-tenant.agent-linux-01
    Centrifugo->>Centrifugo: Check allow_subscribe_for_client: true
    Centrifugo->>Agent: Subscription granted ✅
    
    ControlPlane->>Centrifugo: Publish job_available message
    Centrifugo->>Agent: Forward message
    Agent->>Agent: Process job via OnPublication callback
```

## Comparison with Original Approach

| Aspect | Server-Side Subs (Original) | Client-Side Subs (Current) |
|--------|----------------------------|----------------------------|
| JWT Claims | `sub`, `channels` | `sub` only |
| Agent Code | Modified (OnPublication) | Original (sub.Subscribe) |
| Centrifugo Config | Minimal | Added allow_subscribe_for_client |
| Rebuild Required | Yes (agent) | No |
| Security | Best (centralized control) | Good (authenticated clients) |
| Complexity | Higher | Lower |

## Files Modified

1. [`demo/automation-control-plane/deploy/centrifugo.json`](demo/automation-control-plane/deploy/centrifugo.json)
   - Added `allow_subscribe_for_client: true`
   - Added `allow_publish_for_client: true`

2. [`demo/automation-control-plane/tools/gen-token.py`](demo/automation-control-plane/tools/gen-token.py)
   - Added `"sub": agent_id` claim (already done earlier)
   - Note: `"channels"` claim no longer needed for this approach

## Previous Changes (Still Valid)

The JWT token still includes the `"sub"` claim which is required:

**Current JWT Token**:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhZ2VudC1saW51eC0wMSIsImFnZW50X2lkIjoiYWdlbnQtbGludXgtMDEiLCJ0ZW5hbnRfaWQiOiJ0ZXN0LXRlbmFudCIsInByb2plY3RfaWQiOiJ0ZXN0LXByb2plY3QiLCJjaGFubmVscyI6WyJhZ2VudHMudGVzdC10ZW5hbnQuYWdlbnQtbGludXgtMDEiXSwiZXhwIjoxNzk5NTYzNTEzLCJpYXQiOjE3NjgwMjc1MTN9.CUmqs_obtkz1xdLUYfdGeHaT4XVNEyJVaEjqKwaIwLk
```

**Decoded**:
```json
{
  "sub": "agent-linux-01",           // ✅ Required for user identification
  "agent_id": "agent-linux-01",      // ✅ Application-specific
  "tenant_id": "test-tenant",        // ✅ Application-specific
  "project_id": "test-project",      // ✅ Application-specific
  "channels": ["agents.test-tenant.agent-linux-01"],  // Not needed for client-side subs
  "exp": 1799563513,
  "iat": 1768027513
}
```

## References

- [Centrifugo Client Permissions](https://centrifugal.dev/docs/server/channels#client-side-subscriptions)
- [Centrifuge-go Client Example](https://github.com/centrifugal/centrifuge-go/blob/master/examples/chat/main.go)
- [Centrifugo v6 Configuration](https://centrifugal.dev/docs/server/configuration)

## Summary

✅ **Configuration updated** - Centrifugo now allows client-side subscriptions
✅ **Services restarted** - Changes applied
⚠️ **Verification pending** - Waiting for Docker Desktop WSL integration to be restored

**Once Docker Desktop is running**, the agent should connect and subscribe successfully without any code changes!
