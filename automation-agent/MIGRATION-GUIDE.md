# Migration Guide: JSON to YAML Workflows

This guide helps you migrate from the old JSON workflow format to the new YAML workflow format powered by the probe module.

## Overview

The automation agent now uses the probe task execution framework, which provides:

- More readable YAML format
- Better task organization
- Built-in tasks (HTTP, DB, SSH, Command)
- Enhanced custom tasks (PowerShell, DownloadExec)
- Improved error handling and timeout support

## Key Changes

### 1. Format Change: JSON → YAML

Workflows are now defined in YAML instead of JSON for better readability.

### 2. Structure Changes

**Old JSON Structure**:
```json
{
  "steps": [...]
}
```

**New YAML Structure**:
```yaml
name: workflow-name
timeout: 5m
tasks:
  - name: task-name
    type: task-type
    config:
      ...
```

### 3. Task Type Changes

| Old Type | New Type | Notes |
|----------|----------|-------|
| `exec` | `command` | Renamed for clarity |
| `powershell` | `powershell` | Unchanged |
| `download_exec` | `downloadexec` | Simplified name |

### 4. Configuration Changes

All task-specific configuration now goes in the `config` section.

## Migration Examples

### Example 1: Simple Command Execution

**Before (JSON)**:
```json
{
  "steps": [
    {
      "type": "exec",
      "command": "echo hello world"
    }
  ]
}
```

**After (YAML)**:
```yaml
name: simple-echo
tasks:
  - name: echo-hello
    type: command
    config:
      command: echo hello world
      shell: true
```

### Example 2: PowerShell Script

**Before (JSON)**:
```json
{
  "steps": [
    {
      "type": "powershell",
      "script": "Get-Service | Where-Object {$_.Status -eq 'Running'}"
    }
  ]
}
```

**After (YAML)**:
```yaml
name: list-services
tasks:
  - name: list-running-services
    type: powershell
    config:
      script: |
        Get-Service | Where-Object {$_.Status -eq 'Running'}
      timeout: 30s
```

### Example 3: Download and Execute

**Before (JSON)**:
```json
{
  "steps": [
    {
      "type": "download_exec",
      "artifact": {
        "url": "https://example.com/installer.exe",
        "sha256": "abc123...",
        "signature": "def456...",
        "key_id": "key1"
      }
    }
  ]
}
```

**After (YAML)**:
```yaml
name: install-application
tasks:
  - name: download-and-install
    type: downloadexec
    config:
      url: https://example.com/installer.exe
      sha256: abc123...
      signature: def456...
      public_key: base64_encoded_public_key
      args: ["--silent"]
      timeout: 5m
      cleanup: true
```

**Note**: The signature verification now uses base64-encoded public keys directly instead of key IDs.

### Example 4: Multi-Step Workflow

**Before (JSON)**:
```json
{
  "steps": [
    {
      "type": "exec",
      "command": "git pull origin main"
    },
    {
      "type": "exec",
      "command": "npm install"
    },
    {
      "type": "exec",
      "command": "npm run build"
    }
  ]
}
```

**After (YAML)**:
```yaml
name: deploy-application
timeout: 10m
tasks:
  - name: pull-latest-code
    type: command
    config:
      command: git pull origin main
      shell: true
      timeout: 1m
      
  - name: install-dependencies
    type: command
    config:
      command: npm install
      shell: true
      timeout: 5m
      
  - name: build-application
    type: command
    config:
      command: npm run build
      shell: true
      timeout: 3m
```

## Step-by-Step Migration Process

### Step 1: Identify Your Workflows

List all JSON workflows currently in use:

```sql
SELECT id, workflow FROM jobs WHERE workflow_format = 'json' OR workflow_format IS NULL;
```

### Step 2: Convert Each Workflow

For each workflow:

1. Create the YAML header with name and optional timeout
2. Convert each step to a task with:
   - Unique name
   - Appropriate type
   - Configuration under `config` section
3. Add timeouts where appropriate
4. Add descriptive names for better logging

### Step 3: Test Locally

Before deploying, test each converted workflow:

```bash
# Save your YAML to a file
cat > test-workflow.yaml << 'EOF'
name: test-workflow
tasks:
  - name: test-task
    type: command
    config:
      command: echo "Testing"
      shell: true
EOF

# Test with probe (if you have a test tool)
probe test-workflow.yaml
```

### Step 4: Update Jobs

Update jobs in the control plane database:

```sql
-- Add workflow_format column if it doesn't exist
ALTER TABLE jobs ADD COLUMN workflow_format ENUM('json', 'yaml') DEFAULT 'yaml';

-- Update individual jobs
UPDATE jobs 
SET workflow = '<yaml content here>', 
    workflow_format = 'yaml' 
WHERE id = <job_id>;
```

### Step 5: Deploy Updated Agent

Deploy the new agent version that supports YAML workflows.

### Step 6: Monitor

Monitor the first few workflow executions carefully:

1. Check agent logs for parsing errors
2. Verify task execution results
3. Confirm proper error handling

## Common Migration Issues

### Issue 1: Command Execution

**Problem**: Commands that worked before now fail.

**Solution**: Enable shell mode for complex commands:

```yaml
config:
  command: "ps aux | grep myapp"
  shell: true  # Add this
```

### Issue 2: Multiline Scripts

**Problem**: PowerShell or complex scripts with newlines.

**Solution**: Use YAML multiline syntax:

```yaml
config:
  script: |
    # First line
    # Second line
    # Third line
```

### Issue 3: Signature Verification

**Problem**: Old workflows used key_id, new format uses public_key directly.

**Solution**: Convert key IDs to base64-encoded public keys:

```bash
# If you have the public key file
base64 -w 0 public_key.bin
```

```yaml
config:
  public_key: "bXlwdWJsaWNrZXk..."  # base64-encoded
```

### Issue 4: Timeouts

**Problem**: Tasks timeout that didn't before.

**Solution**: Explicitly set timeouts:

```yaml
config:
  timeout: 5m  # Or appropriate value
```

### Issue 5: Exit Codes

**Problem**: Tasks fail even though they complete successfully.

**Solution**: Ensure scripts exit with code 0 on success:

```yaml
- name: check-something
  type: powershell
  config:
    script: |
      # Your logic here
      if ($success) {
        Write-Output "Success"
        exit 0  # Important!
      } else {
        Write-Error "Failed"
        exit 1
      }
```

## New Capabilities

### HTTP Health Checks

Now built-in:

```yaml
- name: check-api
  type: http
  config:
    url: https://api.example.com/health
    expected_status: [200]
    timeout: 10s
```

### Database Checks

Now built-in:

```yaml
- name: check-database
  type: db
  config:
    driver: mysql
    dsn: user:pass@tcp(localhost:3306)/mydb
    query: SELECT 1
    timeout: 30s
```

### SSH Operations

Now built-in:

```yaml
- name: deploy-via-ssh
  type: ssh
  config:
    host: server.example.com
    user: deploy
    key: /root/.ssh/id_rsa
    command: systemctl restart myapp
```

## Testing Checklist

Before considering migration complete, test:

- [ ] Simple command execution
- [ ] Commands with arguments
- [ ] Shell commands with pipes/redirects
- [ ] PowerShell scripts (Windows)
- [ ] Multi-line scripts
- [ ] Download and execute
- [ ] Signature verification (if used)
- [ ] Timeout handling
- [ ] Error handling
- [ ] Multi-step workflows
- [ ] Workflow-level timeout

## Rollback Plan

If you need to rollback:

1. **Agent**: Redeploy previous agent version
2. **Workflows**: Keep JSON workflows until migration is confirmed successful
3. **Database**: workflow_format column allows both formats during transition

## Support

For migration assistance:

1. Review example workflows in `examples/workflows/`
2. Check probe documentation in `probe/README.md`
3. Test with probe module directly before deploying
4. Start with simple workflows and gradually migrate complex ones

## Dual Format Support (Optional)

If you need to support both formats during transition, modify the agent's message handler:

```go
func (h *MessageHandler) HandleJobAvailable(jobID string) {
    // ... lease job ...
    
    // Detect format
    var results *probe.WorkflowResult
    var err error
    
    if strings.HasPrefix(string(job.Payload), "{") {
        // JSON format (old)
        err = fmt.Errorf("JSON workflows no longer supported")
    } else {
        // YAML format (new)
        results, err = h.probeExecutor.ExecuteYAML(ctx, job.Payload)
    }
    
    // ... handle results ...
}
```

## Timeline Recommendation

**Week 1**: 
- Test probe module with sample workflows
- Convert 2-3 simple workflows

**Week 2**:
- Convert remaining workflows
- Test all converted workflows
- Deploy to dev environment

**Week 3**:
- Monitor dev environment
- Fix any issues discovered
- Deploy to staging environment

**Week 4**:
- Final testing in staging
- Deploy to production
- Monitor closely

**Week 5+**:
- Deprecate JSON format support
- Clean up old workflow code
- Document lessons learned

## Conclusion

The migration to YAML workflows provides:

✅ Better readability and maintainability
✅ More built-in task types
✅ Improved error handling
✅ Better timeout control
✅ Enhanced security (signature verification)
✅ Easier debugging and logging

Take your time with the migration, test thoroughly, and don't hesitate to keep both formats during transition if needed.
