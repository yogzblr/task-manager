#!/usr/bin/env python3
import json
import hmac
import hashlib
import base64
from datetime import datetime, timedelta

def base64url_encode(data):
    """Base64 URL encode without padding"""
    return base64.urlsafe_b64encode(data).rstrip(b'=').decode('utf-8')

def generate_jwt(secret, agent_id, tenant_id, project_id, expiry_days=365):
    """Generate a JWT token"""
    # Header
    header = {
        "alg": "HS256",
        "typ": "JWT"
    }
    
    # Payload
    now = datetime.utcnow()
    exp = now + timedelta(days=expiry_days)
    payload = {
        "agent_id": agent_id,
        "tenant_id": tenant_id,
        "project_id": project_id,
        "exp": int(exp.timestamp()),
        "iat": int(now.timestamp())
    }
    
    # Encode header and payload
    header_encoded = base64url_encode(json.dumps(header, separators=(',', ':')).encode('utf-8'))
    payload_encoded = base64url_encode(json.dumps(payload, separators=(',', ':')).encode('utf-8'))
    
    # Create signature
    message = f"{header_encoded}.{payload_encoded}".encode('utf-8')
    signature = hmac.new(secret.encode('utf-8'), message, hashlib.sha256).digest()
    signature_encoded = base64url_encode(signature)
    
    # Combine to create JWT
    token = f"{header_encoded}.{payload_encoded}.{signature_encoded}"
    return token

if __name__ == "__main__":
    secret = "change-me-in-production"
    agent_id = "agent-linux-01"
    tenant_id = "test-tenant"
    project_id = "test-project"
    
    token = generate_jwt(secret, agent_id, tenant_id, project_id)
    print(token)
