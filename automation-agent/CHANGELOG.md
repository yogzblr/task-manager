# Changelog: Probe Integration

## Version 2.0.0 - Probe Integration

**Release Date**: 2026-01-10

### Major Changes

#### 🎉 New: Probe Task Execution Framework

The agent now uses the probe task execution framework for workflow execution, replacing the previous JSON-based workflow system.

**Benefits**:
- More readable YAML workflow format
- Better task organization and error handling
- Built-in tasks for common scenarios
- Extensible architecture for custom tasks
- Improved timeout and cancellation support

#### 📝 Workflow Format Migration: JSON → YAML

Workflows are now defined in YAML format instead of JSON.

**Before (JSON)**:
```json
{
  "steps": [
    {
      "type": "exec",
      "command": "echo hello"
    }
  ]
}
```

**After (YAML)**:
```yaml
name: my-workflow
tasks:
  - name: echo-hello
    type: command
    config:
      command: echo
      args: ["hello"]
```

### New Features

#### Built-in Task Types

1. **HTTP Task**: HTTP health checks and API testing
   - Validate status codes
   - Custom headers support
   - Configurable timeouts
   
2. **Database Task**: Database connectivity and query execution
   - MySQL support
   - Query result capture
   - Connection testing

3. **SSH Task**: Remote command execution and file transfers
   - Key and password authentication
   - File upload/download
   - Remote command execution

4. **Command Task**: Local shell command execution
   - Direct execution or shell mode
   - Arguments support
   - Cross-platform compatibility

#### Enhanced Custom Tasks

1. **PowerShell Task** (Improved):
   - Better error handling
   - Explicit timeout support
   - Multi-line script support with YAML syntax
   
2. **DownloadExec Task** (Enhanced):
   - Direct public key specification (no key registry needed)
   - Base64-encoded signature support
   - Configurable cleanup
   - Better error messages

### Breaking Changes

#### ⚠️ Workflow Format

- **JSON workflows are no longer supported** by default
- All workflows must be converted to YAML format
- See [MIGRATION-GUIDE.md](MIGRATION-GUIDE.md) for conversion instructions

#### ⚠️ Task Type Names

| Old Name | New Name |
|----------|----------|
| `exec` | `command` |
| `download_exec` | `downloadexec` |

#### ⚠️ Configuration Structure

- All task configuration now goes under a `config` section
- Task-specific fields are no longer at the top level

#### ⚠️ Signature Verification

- `key_id` is replaced with direct `public_key` specification
- Public keys must be base64-encoded Ed25519 keys
- No central key registry required

### Removed Features

#### 🗑️ Deprecated Components

The following internal components have been removed:

- `internal/workflow/` - Old workflow executor
- `internal/plugins/exec/` - Replaced by probe's command task
- `internal/plugins/powershell/` - Replaced by probe's PowerShell task
- `internal/plugins/downloadexec/` - Replaced by probe's DownloadExec task
- `internal/probe/plugin.go` - Old plugin interface
- `internal/probe/task.go` - Old task interface

**Note**: `internal/security/verifier.go` functionality is now integrated into probe's DownloadExec task.

### Dependencies

#### Added

- `github.com/yogzblr/probe v0.0.0` - Task execution framework
- `github.com/go-sql-driver/mysql v1.8.1` - MySQL driver (via probe)
- `golang.org/x/crypto v0.19.0` - SSH and crypto support (via probe)
- `gopkg.in/yaml.v3 v3.0.1` - YAML parsing (via probe)

#### Removed

- Removed dependency on internal workflow package
- Removed dependency on internal plugin packages

### Migration

#### For Users

1. **Convert workflows to YAML format**
   - Use the [MIGRATION-GUIDE.md](MIGRATION-GUIDE.md) for step-by-step instructions
   - Test converted workflows before deploying
   
2. **Update job definitions**
   - Update `workflow_format` field to `yaml`
   - Test job execution with new agent

3. **Update control plane** (optional)
   - Run migration `002_add_workflow_format.sql`
   - Adds `workflow_format` column for tracking

#### For Developers

1. **Agent rebuilding required**
   - Run `go mod tidy` to update dependencies
   - Rebuild agent binary
   
2. **Custom task development**
   - Follow new probe task interface
   - See probe README for custom task documentation

### Documentation

#### New Documentation

- `probe/README.md` - Comprehensive probe framework documentation
- `automation-agent/MIGRATION-GUIDE.md` - JSON to YAML migration guide
- `automation-agent/examples/workflows/` - Example YAML workflows

#### Updated Documentation

- `automation-agent/README.md` - Updated with probe integration details
- Added task reference documentation
- Added troubleshooting section

### Examples

New example workflows added:

1. `simple-health-check.yaml` - HTTP health check
2. `database-check.yaml` - Database connectivity check
3. `windows-deployment.yaml` - Windows application deployment
4. `ssh-deployment.yaml` - Linux deployment via SSH
5. `command-execution.yaml` - Local command execution
6. `mixed-workflow.yaml` - Multi-task comprehensive workflow

### Testing

#### Unit Tests Added

- `probe/probe_test.go` - Core probe functionality tests
- `probe/task_powershell_test.go` - PowerShell task tests
- `probe/task_downloadexec_test.go` - DownloadExec task tests

### Performance

- **Improved**: Task execution overhead reduced
- **Improved**: YAML parsing is faster than JSON for complex workflows
- **Improved**: Better memory usage with streaming task execution

### Security

#### Enhanced

- DownloadExec task now supports direct public key specification
- SHA256 verification is always required for downloaded artifacts
- Ed25519 signature verification is more straightforward

#### Notes

- SSH tasks use `InsecureIgnoreHostKey()` by default
  - **TODO**: Implement proper host key verification in production
- Private keys should have 0600 permissions on Unix systems

### Known Issues

1. **SSH Host Key Verification**: Currently disabled for convenience
   - **Workaround**: Implement proper host key verification for production use
   
2. **Database Driver Support**: Only MySQL is supported
   - **Workaround**: Additional drivers can be added to probe module
   
3. **Parallel Task Execution**: Tasks execute sequentially
   - **Workaround**: Use multiple jobs for parallel execution
   
4. **Windows Path Handling**: PowerShell tasks may have issues with certain path formats
   - **Workaround**: Use forward slashes or escape backslashes

### Upgrade Notes

#### Compatibility

- **Agent**: Version 2.0.0+ required for YAML workflows
- **Control Plane**: No changes required (optional migration available)
- **Workflows**: Must be converted from JSON to YAML

#### Rollback

If rollback is needed:

1. Redeploy agent version 1.x
2. Keep JSON workflows available
3. Revert `workflow_format` column if applied

#### Recommended Upgrade Path

1. **Week 1**: Deploy agent 2.0.0 to dev environment
2. **Week 2**: Convert and test workflows in dev
3. **Week 3**: Deploy to staging and monitor
4. **Week 4**: Deploy to production with careful monitoring

### Contributors

This release integrates the probe framework and extends it with custom tasks for the automation platform.

### License

Proprietary

---

## Previous Versions

### Version 1.x

Initial implementation with JSON workflow support, exec plugin, PowerShell plugin, and DownloadExec plugin.
