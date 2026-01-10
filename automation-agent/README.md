# Automation Agent

Cross-platform automation agent powered by the `probe` task execution framework with control plane integration.

## Features

- **YAML-Based Workflows**: Define automation tasks in readable YAML format
- **Built-in Tasks**: HTTP health checks, database queries, SSH operations, command execution
- **Custom Tasks**: PowerShell scripts (Windows), download-and-execute with verification
- **Signature Verification**: Ed25519 signature verification for downloaded artifacts
- **Control Plane Integration**: Centralized job management and orchestration
- **Multi-Platform**: Supports Windows and Linux
- **Service Support**: Systemd (Linux) and Windows Service integration

## Architecture

The agent integrates the [yogzblr/probe](../probe) task execution framework, which provides:

- Extensible task system
- YAML workflow parsing
- Task execution with timeout and error handling
- Support for HTTP, Database (MySQL), SSH, Command, PowerShell, and DownloadExec tasks

## Installation

### Linux

```bash
sudo ./deploy/linux/install.sh
```

### Windows

```powershell
.\deploy\windows\install.ps1 -ControlPlaneUrl "https://cp.example.com" -TenantId "tenant" -ProjectId "project" -JwtToken "token"
```

## Configuration

Set environment variables or use configuration file:

- `CONTROL_PLANE_URL` - Control plane API URL
- `CENTRIFUGO_URL` - Centrifugo WebSocket URL
- `TENANT_ID` - Tenant ID
- `PROJECT_ID` - Project ID
- `AGENT_ID` - Agent ID (optional, auto-generated if not set)
- `JWT_TOKEN` - JWT authentication token

## Workflow Format

The agent now uses YAML workflows powered by probe. See [examples/workflows/](examples/workflows/) for examples.

### Basic Workflow Structure

```yaml
name: my-workflow
timeout: 5m
tasks:
  - name: task-1
    type: http
    config:
      url: https://example.com
      expected_status: [200]
```

## Available Tasks

### HTTP Task

Performs HTTP requests and health checks.

```yaml
- name: check-api
  type: http
  config:
    url: https://api.example.com/health
    method: GET
    expected_status: [200]
    timeout: 10s
    headers:
      User-Agent: My-Agent/1.0
```

**Configuration**:
- `url` (required): The URL to request
- `method` (optional): HTTP method (default: GET)
- `expected_status` (optional): Expected status codes (default: [200])
- `timeout` (optional): Request timeout (default: 30s)
- `headers` (optional): Custom HTTP headers

### Database Task

Executes database queries and checks connectivity.

```yaml
- name: check-database
  type: db
  config:
    driver: mysql
    dsn: user:password@tcp(localhost:3306)/mydb
    query: SELECT COUNT(*) FROM users
    timeout: 30s
```

**Configuration**:
- `driver` (required): Database driver (currently supports: mysql)
- `dsn` (required): Data source name (connection string)
- `query` (required): SQL query to execute
- `timeout` (optional): Query timeout (default: 30s)

### SSH Task

Executes commands or transfers files via SSH.

```yaml
- name: deploy-config
  type: ssh
  config:
    host: server.example.com
    port: 22
    user: deploy
    key: /root/.ssh/id_rsa
    # For file upload:
    upload:
      local: /tmp/config.yaml
      remote: /etc/app/config.yaml
    # For command execution:
    command: systemctl restart myapp
    timeout: 60s
```

**Configuration**:
- `host` (required): SSH server hostname
- `port` (optional): SSH port (default: 22)
- `user` (required): SSH username
- `key` or `password` (required): Authentication method
- `command` (optional): Command to execute
- `upload` (optional): File upload configuration
- `timeout` (optional): Operation timeout (default: 60s)

### Command Task

Executes local shell commands.

```yaml
- name: backup-database
  type: command
  config:
    command: mysqldump
    args: ["-u", "root", "mydb"]
    # Or use shell mode:
    # command: "mysqldump -u root mydb > backup.sql"
    # shell: true
    timeout: 5m
```

**Configuration**:
- `command` (required): Command to execute
- `args` (optional): Command arguments (if shell is false)
- `shell` (optional): Execute through shell (default: false)
- `timeout` (optional): Execution timeout (default: 30s)

### PowerShell Task (Windows Only)

Executes PowerShell scripts on Windows.

```yaml
- name: check-service
  type: powershell
  config:
    script: |
      $service = Get-Service -Name "MyService"
      if ($service.Status -eq "Running") {
        Write-Output "Service is running"
        exit 0
      } else {
        Write-Error "Service is not running"
        exit 1
      }
    timeout: 30s
```

**Configuration**:
- `script` (required): PowerShell script to execute
- `timeout` (optional): Execution timeout (default: 30s)

**Note**: This task will fail if executed on non-Windows platforms.

### DownloadExec Task

Downloads a file from a URL, verifies its integrity, and executes it.

```yaml
- name: run-installer
  type: downloadexec
  config:
    url: https://releases.example.com/installer.exe
    sha256: abc123def456...  # Required
    signature: base64_encoded_signature  # Optional
    public_key: base64_encoded_ed25519_public_key  # Required if signature provided
    args: ["--silent", "--install-dir=/opt/app"]
    timeout: 5m
    cleanup: true
```

**Configuration**:
- `url` (required): Download URL
- `sha256` (required): Expected SHA256 hash of the file
- `signature` (optional): Base64-encoded Ed25519 signature
- `public_key` (required if signature provided): Base64-encoded Ed25519 public key
- `args` (optional): Command line arguments to pass to the executable
- `timeout` (optional): Execution timeout (default: 60s)
- `cleanup` (optional): Delete file after execution (default: true)

**Security**: 
- SHA256 verification is always required
- Ed25519 signature verification is optional but recommended
- Files are downloaded to temporary locations
- Automatic cleanup after execution (configurable)

## Example Workflows

See [examples/workflows/](examples/workflows/) for complete examples:

- **simple-health-check.yaml**: HTTP health check example
- **database-check.yaml**: Database connectivity and query examples
- **windows-deployment.yaml**: Windows application deployment with PowerShell
- **ssh-deployment.yaml**: Linux deployment via SSH
- **command-execution.yaml**: Local command execution examples
- **mixed-workflow.yaml**: Multi-task workflow combining different task types

## Migration from JSON Workflows

The agent previously used JSON-based workflows. To migrate:

1. Convert workflow structure to YAML format
2. Update task type names if needed
3. Move inline configuration to the `config` section
4. Test workflows locally before deploying

### Example Migration

**Old JSON format**:
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

**New YAML format**:
```yaml
name: my-workflow
tasks:
  - name: echo-hello
    type: command
    config:
      command: echo
      args: ["hello"]
```

## Development

### Building

```bash
cd automation-agent
go build -o automation-agent ./cmd/agent
```

### Testing

```bash
go test ./...
```

### Adding Custom Tasks

To add custom tasks, extend the probe module in the `probe/` directory. See the probe README for details.

## Troubleshooting

### Agent won't start

1. Check environment variables are set correctly
2. Verify network connectivity to control plane
3. Check JWT token is valid
4. Review logs in system journal (Linux) or Event Viewer (Windows)

### Workflow execution fails

1. Validate YAML syntax
2. Check task configuration parameters
3. Verify agent has necessary permissions
4. Check task-specific requirements (e.g., PowerShell on Windows)

### SSH tasks fail

1. Verify SSH key permissions (0600 on Linux)
2. Check SSH server is accessible
3. Verify user has necessary permissions on remote server
4. Test SSH connection manually: `ssh -i /path/to/key user@host`

## License

Proprietary
