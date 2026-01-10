# Probe Quick Start Guide

Get started with probe in 5 minutes.

## Installation

```bash
go get github.com/yogzblr/probe
```

## Your First Workflow

Create a file `hello.yaml`:

```yaml
name: hello-world
tasks:
  - name: say-hello
    type: command
    config:
      command: echo
      args: ["Hello, Probe!"]
```

## Run It

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/yogzblr/probe"
)

func main() {
    // Read workflow file
    data, err := os.ReadFile("hello.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create probe and execute
    p := probe.New()
    result, err := p.ExecuteYAML(context.Background(), data)
    if err != nil {
        log.Fatalf("Execution failed: %v", err)
    }
    
    // Print results
    fmt.Printf("Workflow: %s\n", result.Name)
    fmt.Printf("Success: %v\n", result.Success)
    
    for _, task := range result.Tasks {
        fmt.Printf("\nTask: %s\n", task.Name)
        fmt.Printf("  Success: %v\n", task.Success)
        if task.Output != nil {
            fmt.Printf("  Output: %v\n", task.Output)
        }
    }
}
```

## Common Tasks

### HTTP Health Check

```yaml
name: health-check
tasks:
  - name: check-api
    type: http
    config:
      url: https://api.github.com
      expected_status: [200]
```

### Database Query

```yaml
name: db-check
tasks:
  - name: query-database
    type: db
    config:
      driver: mysql
      dsn: user:pass@tcp(localhost:3306)/mydb
      query: SELECT VERSION()
```

### SSH Command

```yaml
name: remote-command
tasks:
  - name: check-uptime
    type: ssh
    config:
      host: example.com
      user: ubuntu
      key: ~/.ssh/id_rsa
      command: uptime
```

### PowerShell (Windows)

```yaml
name: windows-check
tasks:
  - name: list-services
    type: powershell
    config:
      script: |
        Get-Service | Where-Object {$_.Status -eq 'Running'} | Select-Object -First 5
```

### Download & Execute

```yaml
name: install-tool
tasks:
  - name: download-and-run
    type: downloadexec
    config:
      url: https://example.com/tool.sh
      sha256: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
      args: ["--install"]
```

## Multi-Task Workflow

```yaml
name: deployment
timeout: 10m
tasks:
  # Check API health
  - name: pre-check
    type: http
    config:
      url: https://api.example.com/health
      expected_status: [200]
  
  # Deploy via SSH
  - name: deploy
    type: ssh
    config:
      host: server.example.com
      user: deploy
      key: ~/.ssh/deploy_key
      command: |
        cd /app &&
        git pull &&
        docker-compose restart
  
  # Verify deployment
  - name: post-check
    type: http
    config:
      url: https://api.example.com/version
      expected_status: [200]
```

## Error Handling

```go
result, err := p.ExecuteYAML(ctx, data)
if err != nil {
    fmt.Printf("Workflow failed: %v\n", err)
    
    // Check which tasks succeeded
    for _, task := range result.Tasks {
        if task.Success {
            fmt.Printf("✓ %s succeeded\n", task.Name)
        } else {
            fmt.Printf("✗ %s failed: %s\n", task.Name, task.Error)
        }
    }
}
```

## Timeouts

### Task-Level Timeout

```yaml
- name: slow-task
  type: command
  config:
    command: sleep 5
    timeout: 10s  # Task timeout
```

### Workflow-Level Timeout

```yaml
name: time-limited-workflow
timeout: 5m  # Entire workflow must complete within 5 minutes
tasks:
  - name: task-1
    type: command
    config:
      command: echo "Step 1"
  - name: task-2
    type: command
    config:
      command: echo "Step 2"
```

### Context Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
defer cancel()

result, err := p.ExecuteYAML(ctx, data)
```

## Custom Tasks

```go
// Define your task
type MyTask struct {
    URL string
}

func (t *MyTask) Configure(config map[string]interface{}) error {
    t.URL = config["url"].(string)
    return nil
}

func (t *MyTask) Execute(ctx context.Context) (interface{}, error) {
    // Your logic here
    return map[string]interface{}{
        "result": "success",
    }, nil
}

// Register it
p := probe.New()
p.RegisterTask("mytask", func() probe.Task {
    return &MyTask{}
})

// Use it
yaml := `
name: custom-workflow
tasks:
  - name: my-custom-task
    type: mytask
    config:
      url: https://example.com
`
```

## Best Practices

### 1. Always Set Timeouts

```yaml
# Good
- name: api-call
  type: http
  config:
    url: https://api.example.com
    timeout: 30s

# Bad - uses default, may be too long
- name: api-call
  type: http
  config:
    url: https://api.example.com
```

### 2. Give Tasks Descriptive Names

```yaml
# Good
- name: check-database-connectivity
  type: db
  config:
    driver: mysql
    dsn: ...
    query: SELECT 1

# Bad
- name: task1
  type: db
  config:
    ...
```

### 3. Use Shell Mode for Complex Commands

```yaml
# Good for pipes and redirects
- name: backup
  type: command
  config:
    command: "mysqldump mydb | gzip > backup.sql.gz"
    shell: true

# Good for simple commands
- name: list-files
  type: command
  config:
    command: ls
    args: ["-la", "/tmp"]
    shell: false
```

### 4. Always Verify Downloaded Files

```yaml
# Good - includes SHA256
- name: download
  type: downloadexec
  config:
    url: https://example.com/file
    sha256: abc123...

# Even better - includes signature
- name: download
  type: downloadexec
  config:
    url: https://example.com/file
    sha256: abc123...
    signature: def456...
    public_key: base64key...
```

### 5. Use Multiline Syntax for Scripts

```yaml
# Good - readable
- name: complex-script
  type: powershell
  config:
    script: |
      $services = Get-Service
      foreach ($service in $services) {
        Write-Output $service.Name
      }

# Bad - hard to read
- name: complex-script
  type: powershell
  config:
    script: "$services = Get-Service; foreach ($service in $services) { Write-Output $service.Name }"
```

## Common Patterns

### Health Check Pattern

```yaml
name: comprehensive-health-check
tasks:
  - name: check-api
    type: http
    config:
      url: https://api.example.com/health
      expected_status: [200]
      
  - name: check-database
    type: db
    config:
      driver: mysql
      dsn: user:pass@tcp(db:3306)/app
      query: SELECT 1
      
  - name: check-disk-space
    type: command
    config:
      command: df -h /
      shell: true
```

### Deployment Pattern

```yaml
name: deploy-application
tasks:
  - name: pull-code
    type: ssh
    config:
      host: server.example.com
      user: deploy
      key: ~/.ssh/deploy_key
      command: cd /app && git pull origin main
      
  - name: build
    type: ssh
    config:
      host: server.example.com
      user: deploy
      key: ~/.ssh/deploy_key
      command: cd /app && docker-compose build
      
  - name: restart
    type: ssh
    config:
      host: server.example.com
      user: deploy
      key: ~/.ssh/deploy_key
      command: cd /app && docker-compose up -d
      
  - name: verify
    type: http
    config:
      url: https://app.example.com/health
      expected_status: [200]
```

### Backup Pattern

```yaml
name: backup-database
tasks:
  - name: create-backup
    type: command
    config:
      command: "mysqldump -u root mydb | gzip > /backup/mydb-$(date +%Y%m%d).sql.gz"
      shell: true
      timeout: 10m
      
  - name: verify-backup
    type: command
    config:
      command: "test -f /backup/mydb-$(date +%Y%m%d).sql.gz && echo 'Backup created'"
      shell: true
      
  - name: upload-to-remote
    type: ssh
    config:
      host: backup-server.example.com
      user: backup
      key: ~/.ssh/backup_key
      upload:
        local: /backup/mydb-20260110.sql.gz
        remote: /backups/mydb-20260110.sql.gz
```

## Debugging

### Enable Verbose Output

```go
result, err := p.ExecuteYAML(ctx, data)

// Print detailed task info
for i, task := range result.Tasks {
    fmt.Printf("Task %d: %s (%s)\n", i+1, task.Name, task.Type)
    fmt.Printf("  Success: %v\n", task.Success)
    if !task.Success {
        fmt.Printf("  Error: %s\n", task.Error)
    }
    if task.Output != nil {
        fmt.Printf("  Output: %+v\n", task.Output)
    }
}
```

### Test Individual Tasks

```yaml
# Test just one task
name: test-single-task
tasks:
  - name: test-command
    type: command
    config:
      command: echo
      args: ["Testing"]
```

## Next Steps

1. Read the [full README](README.md) for detailed documentation
2. Explore [example workflows](../automation-agent/examples/workflows/)
3. Check [migration guide](../automation-agent/MIGRATION-GUIDE.md) if coming from JSON workflows
4. Build your own custom tasks

## Getting Help

- Check probe module documentation: `probe/README.md`
- Review example workflows: `automation-agent/examples/workflows/`
- See troubleshooting section in agent README

## Resources

- [Full Probe Documentation](README.md)
- [Agent Integration Guide](../automation-agent/README.md)
- [Migration Guide](../automation-agent/MIGRATION-GUIDE.md)
- [Example Workflows](../automation-agent/examples/workflows/)
