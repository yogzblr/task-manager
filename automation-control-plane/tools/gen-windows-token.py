import jwt
import time

secret = "change-me-in-production"
agent_id = "agent-windows-01"
tenant_id = "test-tenant"
project_id = "test-project"

payload = {
    "agent_id": agent_id,
    "tenant_id": tenant_id,
    "project_id": project_id,
    "exp": int(time.time()) + (365 * 24 * 60 * 60),
    "iat": int(time.time())
}

token = jwt.encode(payload, secret, algorithm="HS256")
print(token)
