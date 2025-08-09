package ssh

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	client *ssh.Client
	config *ssh.ClientConfig
}

type KeyPair struct {
	PrivateKey string
	PublicKey  string
	KeyName    string
}

// GenerateKeyPair creates a new RSA key pair for SSH access
func GenerateKeyPair(keyName string) (*KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generating private key: %w", err)
	}

	// Encode private key to PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyBytes := pem.EncodeToMemory(privateKeyPEM)

	// Generate public key in SSH format
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("generating public key: %w", err)
	}

	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)

	return &KeyPair{
		PrivateKey: string(privateKeyBytes),
		PublicKey:  strings.TrimSpace(string(publicKeyBytes)),
		KeyName:    keyName,
	}, nil
}

// SaveKeyPairToFile saves the key pair to local files
func SaveKeyPairToFile(keyPair *KeyPair, privateKeyPath string) error {
	// Save private key
	err := os.WriteFile(privateKeyPath, []byte(keyPair.PrivateKey), 0600)
	if err != nil {
		return fmt.Errorf("saving private key: %w", err)
	}

	// Save public key
	publicKeyPath := privateKeyPath + ".pub"
	err = os.WriteFile(publicKeyPath, []byte(keyPair.PublicKey), 0644)
	if err != nil {
		return fmt.Errorf("saving public key: %w", err)
	}

	return nil
}

// NewClient creates a new SSH client
func NewClient(host, user, privateKeyPath string) (*Client, error) {
	// Read private key
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}

	// Parse private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}

	// SSH client configuration
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For development
		Timeout:         30 * time.Second,
	}

	return &Client{
		config: config,
	}, nil
}

// Connect establishes SSH connection to the host
func (c *Client) Connect(ctx context.Context, host string) error {
	// Add default SSH port if not specified
	if !strings.Contains(host, ":") {
		host = host + ":22"
	}

	// Connect with context timeout
	connectTimeout := 30 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		if until := time.Until(deadline); until < connectTimeout {
			connectTimeout = until
		}
	}

	// Create connection with timeout
	conn, err := net.DialTimeout("tcp", host, connectTimeout)
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", host, err)
	}

	// Create SSH connection
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, host, c.config)
	if err != nil {
		conn.Close()
		return fmt.Errorf("SSH handshake: %w", err)
	}

	c.client = ssh.NewClient(sshConn, chans, reqs)
	return nil
}

// WaitForConnection waits until SSH connection is available
func (c *Client) WaitForConnection(ctx context.Context, host string, maxRetries int) error {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := c.Connect(ctx, host)
		if err == nil {
			return nil
		}

		lastErr = err
		fmt.Printf("SSH connection attempt %d/%d failed: %v\n", i+1, maxRetries, err)
		
		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}

	return fmt.Errorf("failed to establish SSH connection after %d attempts: %w", maxRetries, lastErr)
}

// ExecuteCommand runs a command over SSH
func (c *Client) ExecuteCommand(ctx context.Context, command string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("SSH client not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("creating session: %w", err)
	}
	defer session.Close()

	// Set up output capture
	var output strings.Builder
	session.Stdout = &output
	session.Stderr = &output

	// Execute command with context
	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()

	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGKILL)
		return "", ctx.Err()
	case err := <-done:
		if err != nil {
			return output.String(), fmt.Errorf("command failed: %w", err)
		}
		return output.String(), nil
	}
}

// ExecuteCommandStream runs a command and streams output
func (c *Client) ExecuteCommandStream(ctx context.Context, command string, stdout, stderr io.Writer) error {
	if c.client == nil {
		return fmt.Errorf("SSH client not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}
	defer session.Close()

	// Set up output streams
	session.Stdout = stdout
	session.Stderr = stderr

	// Execute command with context
	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()

	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGKILL)
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// UploadFile uploads a file via SCP-like functionality
func (c *Client) UploadFile(ctx context.Context, localPath, remotePath string) error {
	if c.client == nil {
		return fmt.Errorf("SSH client not connected")
	}

	// Read local file
	content, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("reading local file: %w", err)
	}

	// Create file on remote system
	command := fmt.Sprintf("cat > %s", remotePath)
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}
	defer session.Close()

	session.Stdin = strings.NewReader(string(content))

	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()

	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGKILL)
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// TestConnection tests if SSH connection is working
func (c *Client) TestConnection(ctx context.Context) error {
	output, err := c.ExecuteCommand(ctx, "echo 'SSH connection successful'")
	if err != nil {
		return fmt.Errorf("test command failed: %w", err)
	}

	if !strings.Contains(output, "SSH connection successful") {
		return fmt.Errorf("unexpected test output: %s", output)
	}

	return nil
}