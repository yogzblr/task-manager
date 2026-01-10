package probe

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHTask performs SSH operations
type SSHTask struct {
	Host     string
	Port     int
	User     string
	Key      string
	Password string
	Command  string
	Upload   *SSHUpload
	Timeout  time.Duration
}

// SSHUpload represents a file upload configuration
type SSHUpload struct {
	Local  string
	Remote string
}

// Configure sets up the SSH task
func (t *SSHTask) Configure(config map[string]interface{}) error {
	// Host is required
	host, ok := config["host"].(string)
	if !ok || host == "" {
		return fmt.Errorf("host is required")
	}
	t.Host = host
	
	// Port (default: 22)
	if port, ok := config["port"].(int); ok {
		t.Port = port
	} else {
		t.Port = 22
	}
	
	// User is required
	user, ok := config["user"].(string)
	if !ok || user == "" {
		return fmt.Errorf("user is required")
	}
	t.User = user
	
	// Authentication: key or password
	if key, ok := config["key"].(string); ok {
		t.Key = key
	}
	if password, ok := config["password"].(string); ok {
		t.Password = password
	}
	
	if t.Key == "" && t.Password == "" {
		return fmt.Errorf("either key or password is required")
	}
	
	// Command (optional)
	if command, ok := config["command"].(string); ok {
		t.Command = command
	}
	
	// Upload (optional)
	if upload, ok := config["upload"].(map[string]interface{}); ok {
		local, _ := upload["local"].(string)
		remote, _ := upload["remote"].(string)
		if local != "" && remote != "" {
			t.Upload = &SSHUpload{
				Local:  local,
				Remote: remote,
			}
		}
	}
	
	// Timeout (default: 60s)
	if timeoutStr, ok := config["timeout"].(string); ok {
		duration, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout: %w", err)
		}
		t.Timeout = duration
	} else {
		t.Timeout = 60 * time.Second
	}
	
	return nil
}

// Execute performs the SSH operation
func (t *SSHTask) Execute(ctx context.Context) (interface{}, error) {
	// Create SSH client config
	config := &ssh.ClientConfig{
		User:            t.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // WARNING: insecure, use proper host key verification in production
		Timeout:         t.Timeout,
	}
	
	// Add authentication method
	if t.Key != "" {
		key, err := os.ReadFile(t.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}
		
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else if t.Password != "" {
		config.Auth = []ssh.AuthMethod{ssh.Password(t.Password)}
	}
	
	// Connect to SSH server
	addr := fmt.Sprintf("%s:%d", t.Host, t.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %w", err)
	}
	defer client.Close()
	
	result := make(map[string]interface{})
	
	// Handle file upload if specified
	if t.Upload != nil {
		if err := t.uploadFile(client); err != nil {
			return nil, fmt.Errorf("file upload failed: %w", err)
		}
		result["upload"] = "success"
	}
	
	// Execute command if specified
	if t.Command != "" {
		output, err := t.executeCommand(client)
		if err != nil {
			return nil, fmt.Errorf("command execution failed: %w", err)
		}
		result["output"] = output
	}
	
	return result, nil
}

func (t *SSHTask) executeCommand(client *ssh.Client) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()
	
	output, err := session.CombinedOutput(t.Command)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}
	
	return string(output), nil
}

func (t *SSHTask) uploadFile(client *ssh.Client) error {
	// Open local file
	localFile, err := os.Open(t.Upload.Local)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()
	
	// Get file info
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local file: %w", err)
	}
	
	// Create session
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()
	
	// Create stdin pipe
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	
	// Start SCP command
	go func() {
		defer stdin.Close()
		fmt.Fprintf(stdin, "C0644 %d %s\n", fileInfo.Size(), fileInfo.Name())
		io.Copy(stdin, localFile)
		fmt.Fprint(stdin, "\x00")
	}()
	
	// Run SCP command
	if err := session.Run(fmt.Sprintf("scp -t %s", t.Upload.Remote)); err != nil {
		return fmt.Errorf("scp command failed: %w", err)
	}
	
	return nil
}
