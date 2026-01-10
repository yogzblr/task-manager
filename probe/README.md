# Probe - Task Execution Framework

Forked and extended from linyows/probe with custom tasks for the automation platform.

## Overview

Probe is a flexible task execution framework that uses YAML-based workflow definitions. It provides a set of built-in tasks for common automation scenarios and allows easy extension with custom tasks.

## Features

- **YAML-Based Workflows**: Define complex automation workflows in readable YAML format
- **Built-in Tasks**: HTTP, Database (MySQL), SSH, Command execution
- **Custom Tasks**: PowerShell (Windows-only), DownloadExec with signature verification
- **Extensible Architecture**: Easy to add new task types
- **Context-Aware**: Proper timeout and cancellation support
- **Type-Safe**: Strongly-typed task configuration

## Installation

```bash
go get github.com/yogzblr/probe
```

## Usage

### Basic Example

```go
package main

import (
    "context"
    "log"
    
    "github.com/yogzblr/probe"
)

func main() {
    // Create probe instance
    p := probe.New()
    
    // Define workflow YAML
    yaml := `
name: health-check
tasks:
  - name: check-api
    type: http
    config:
      url: https://api.example.com/health
      expected_status: [200]
`
    
    // Execute workflow
    result, err := p.ExecuteYAML(context.Background(), []byte(yaml))
    if err != nil {
        log.Fatalf("Execution failed: %v", err)
    }
    
    log.Printf("Workflow %s completed successfully: %v", result.Name, result.Success)
}
```

### Programmatic Workflow Definition

```go
workflow := &probe.Workflow{
    Name: "my-workflow",
    Tasks: []probe.TaskDefinition{
        {
            Name: "check-database",
            Type: "db",
            Config: map[string]interface{}{
                "driver": "mysql",
                "dsn":    "user:pass@tcp(localhost:3306)/mydb",
                "query":  "SELECT 1",
            },
        },
    },
}

result, err := p.Execute(context.Background(), workflow)
```

## Built-in Tasks

### HTTP Task

Performs HTTP requests and validates responses.

```yaml
- name: api-check
  type: http
  config:
    url: https://api.example.com/health
    method: GET
    expected_status: [200, 201]
    timeout: 10s
    headers:
      Authorization: Bearer token123
```

**Parameters**:
- `url` (string, required): Target URL
- `method` (string, optional): HTTP method (default: GET)
- `expected_status` ([]int, optional): Expected status codes (default: [200])
- `timeout` (string, optional): Request timeout (default: 30s)
- `headers` (map, optional): Custom HTTP headers

### Database Task

Executes SQL queries and checks database connectivity.

```yaml
- name: db-check
  type: db
  config:
    driver: mysql
    dsn: user:password@tcp(localhost:3306)/database
    query: SELECT COUNT(*) FROM users WHERE active = 1
    timeout: 30s
```

**Parameters**:
- `driver` (string, required): Database driver (currently: mysql)
- `dsn` (string, required): Data source name / connection string
- `query` (string, required): SQL query to execute
- `timeout` (string, optional): Query timeout (default: 30s)

**Supported Drivers**:
- `mysql`: MySQL/MariaDB

### SSH Task

Executes commands or transfers files over SSH.

```yaml
- name: deploy-file
  type: ssh
  config:
    host: server.example.com
    port: 22
    user: deploy
    key: /path/to/private/key
    upload:
      local: /local/config.yaml
      remote: /remote/config.yaml
    timeout: 60s
```

```yaml
- name: restart-service
  type: ssh
  config:
    host: server.example.com
    user: deploy
    key: /path/to/private/key
    command: sudo systemctl restart myapp
```

**Parameters**:
- `host` (string, required): SSH server hostname/IP
- `port` (int, optional): SSH port (default: 22)
- `user` (string, required): SSH username
- `key` (string): Path to private key file
- `password` (string): Password authentication (alternative to key)
- `command` (string, optional): Command to execute
- `upload` (object, optional): File upload configuration
  - `local` (string): Local file path
  - `remote` (string): Remote file path
- `timeout` (string, optional): Operation timeout (default: 60s)

**Note**: Either `key` or `password` must be provided for authentication.

### Command Task

Executes local shell commands.

```yaml
- name: backup
  type: command
  config:
    command: tar
    args: ["-czf", "/backup/app.tar.gz", "/var/app"]
    timeout: 5m
```

```yaml
- name: shell-script
  type: command
  config:
    command: "ps aux | grep myapp | wc -l"
    shell: true
    timeout: 10s
```

**Parameters**:
- `command` (string, required): Command to execute
- `args` ([]string, optional): Command arguments (used when shell=false)
- `shell` (bool, optional): Execute through shell (default: false)
- `timeout` (string, optional): Execution timeout (default: 30s)

**Shell Mode**:
- When `shell: false`: Executes command directly with args
- When `shell: true`: Executes command through system shell (cmd.exe on Windows, /bin/sh on Unix)

## Custom Tasks

### PowerShell Task (Windows Only)

Executes PowerShell scripts. Only available on Windows systems.

```yaml
- name: check-windows-service
  type: powershell
  config:
    script: |
      $service = Get-Service -Name "W3SVC"
      if ($service.Status -eq "Running") {
        Write-Output "IIS is running"
        exit 0
      } else {
        Write-Error "IIS is not running"
        exit 1
      }
    timeout: 30s
```

**Parameters**:
- `script` (string, required): PowerShell script content
- `timeout` (string, optional): Execution timeout (default: 30s)

**Platform**: Windows only. Task will fail with an error on non-Windows platforms.

### DownloadExec Task

Downloads a file, verifies its integrity, and executes it.

```yaml
- name: install-app
  type: downloadexec
  config:
    url: https://releases.example.com/app-v1.2.3.tar.gz
    sha256: a3b2c1d4e5f6...  # Full SHA256 hash
    signature: base64_signature  # Optional Ed25519 signature
    public_key: base64_public_key  # Required if signature provided
    args: ["--install", "/opt/app"]
    timeout: 5m
    cleanup: true
```

**Parameters**:
- `url` (string, required): Download URL
- `sha256` (string, required): Expected SHA256 hash (hex encoded)
- `signature` (string, optional): Base64-encoded Ed25519 signature
- `public_key` (string, required if signature): Base64-encoded Ed25519 public key (32 bytes)
- `args` ([]string, optional): Arguments to pass to executable
- `timeout` (string, optional): Execution timeout (default: 60s)
- `cleanup` (bool, optional): Delete file after execution (default: true)

**Security Features**:
1. **SHA256 Verification** (Required): Ensures file integrity
2. **Ed25519 Signature Verification** (Optional): Ensures file authenticity
3. **Temporary Files**: Downloads to secure temporary location
4. **Automatic Cleanup**: Removes files after execution (configurable)

**Example with Signature Verification**:

```go
// Generate signature (offline, during build):
// 1. Generate Ed25519 key pair
// 2. Sign the file: signature = sign(fileContent, privateKey)
// 3. Encode: base64(signature)

// In workflow:
- name: secure-install
  type: downloadexec
  config:
    url: https://releases.example.com/app.exe
    sha256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    signature: "SGVsbG8gV29ybGQh..."
    public_key: "bXlwdWJsaWNrZXk="
    cleanup: true
```

## Extending Probe

### Creating a Custom Task

```go
package mytasks

import (
    "context"
    "github.com/yogzblr/probe"
)

// MyCustomTask implements probe.Task
type MyCustomTask struct {
    // Configuration fields
    Parameter1 string
    Parameter2 int
}

func (t *MyCustomTask) Configure(config map[string]interface{}) error {
    // Parse and validate configuration
    if param1, ok := config["parameter1"].(string); ok {
        t.Parameter1 = param1
    } else {
        return fmt.Errorf("parameter1 is required")
    }
    
    if param2, ok := config["parameter2"].(int); ok {
        t.Parameter2 = param2
    }
    
    return nil
}

func (t *MyCustomTask) Execute(ctx context.Context) (interface{}, error) {
    // Implement task logic
    result := map[string]interface{}{
        "output": "Task completed successfully",
    }
    return result, nil
}
```

### Registering Custom Tasks

```go
p := probe.New()

// Register custom task
p.RegisterTask("mycustom", func() probe.Task {
    return &MyCustomTask{}
})

// Use in workflow
yaml := `
name: custom-workflow
tasks:
  - name: my-task
    type: mycustom
    config:
      parameter1: value1
      parameter2: 42
`

result, err := p.ExecuteYAML(context.Background(), []byte(yaml))
```

## Workflow Structure

### Complete Workflow Example

```yaml
name: comprehensive-deployment
timeout: 10m
tasks:
  # Step 1: Check prerequisites
  - name: check-api-health
    type: http
    config:
      url: https://api.example.com/health
      expected_status: [200]
      
  # Step 2: Verify database
  - name: check-database
    type: db
    config:
      driver: mysql
      dsn: user:pass@tcp(db:3306)/app
      query: SELECT 1
      
  # Step 3: Deploy to server
  - name: upload-artifacts
    type: ssh
    config:
      host: app-server.example.com
      user: deploy
      key: /keys/deploy_key
      upload:
        local: /build/app.tar.gz
        remote: /tmp/app.tar.gz
        
  # Step 4: Extract and restart
  - name: deploy-app
    type: ssh
    config:
      host: app-server.example.com
      user: deploy
      key: /keys/deploy_key
      command: |
        cd /opt/app &&
        tar xzf /tmp/app.tar.gz &&
        sudo systemctl restart app
        
  # Step 5: Verify deployment
  - name: verify-deployment
    type: http
    config:
      url: https://app-server.example.com/version
      expected_status: [200]
```

### Workflow Fields

- `name` (string, required): Workflow name
- `timeout` (string, optional): Global workflow timeout
- `tasks` ([]TaskDefinition, required): List of tasks to execute

### Task Definition Fields

- `name` (string, required): Task name (for logging and identification)
- `type` (string, required): Task type (http, db, ssh, command, powershell, downloadexec)
- `config` (map, required): Task-specific configuration

## Error Handling

Probe uses a fail-fast approach:

1. If a task fails, workflow execution stops immediately
2. All completed task results are returned
3. Error information is included in the result

```go
result, err := p.ExecuteYAML(ctx, yamlData)
if err != nil {
    // Workflow failed
    log.Printf("Workflow failed: %v", err)
    
    // Check partial results
    for _, taskResult := range result.Tasks {
        if taskResult.Success {
            log.Printf("Task %s succeeded", taskResult.Name)
        } else {
            log.Printf("Task %s failed: %s", taskResult.Name, taskResult.Error)
        }
    }
}
```

## Testing

```bash
go test ./...
```

Run specific tests:

```bash
go test -v -run TestProbeHTTPTask
go test -v -run TestPowerShellTask  # Windows only
```

## Performance Considerations

- Tasks execute sequentially (no parallel execution within a workflow)
- Each task respects context cancellation
- Timeouts are enforced at task level
- Database connections are created per task (connection pooling in DSN)
- SSH connections are established per task

## Security Best Practices

1. **DownloadExec Task**:
   - Always use SHA256 verification
   - Use signature verification for production deployments
   - Keep private keys secure and separate from workflows
   - Use HTTPS URLs for downloads

2. **SSH Task**:
   - Use key-based authentication over passwords
   - Set proper file permissions on private keys (0600)
   - Use dedicated deployment keys with limited privileges
   - Consider SSH certificate authentication

3. **Database Task**:
   - Use read-only database users for health checks
   - Don't include credentials in workflow YAML (use environment variables)
   - Consider network-level access controls

4. **PowerShell Task**:
   - Validate and sanitize any dynamic input
   - Run agent with minimal necessary privileges
   - Use PowerShell execution policies appropriately

## License

MIT License
