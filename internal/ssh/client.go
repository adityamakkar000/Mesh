package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	conn *ssh.Client
}

func Connect(ctx context.Context, user, host, identityFile string) (*Client, error) {
	key, err := os.ReadFile(identityFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read identity file: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:22", host))
	if err != nil {
		return nil, err
	}

	// ssh.Dial is not safe, we use DialContext instead
	c, chans, reqs, err := ssh.NewClientConn(conn, fmt.Sprintf("%s:22", host), config)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{conn: ssh.NewClient(c, chans, reqs)}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Synchronous execution of a command (setup only)
func (c *Client) Exec(ctx context.Context, command string, stdout, stderr io.Writer) error {
	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr

	errCh := make(chan error, 1)

	go func() {
		errCh <- session.Run(command)
	}()

	select {
	case err := <-errCh:
		return err

	case <-ctx.Done():
		_ = session.Signal(ssh.SIGTERM)
		select {
		case <-errCh:
		case <-time.After(2 * time.Second):
		}
		return ctx.Err()
	}
}

// Asynchronous execution of a command (exec-and-forget)
func (c *Client) ExecDetached(ctx context.Context, command string) error {
	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	wrapped := fmt.Sprintf("setsid %s > /dev/null 2>&1 < /dev/null &", command)

	errCh := make(chan error, 1)
	go func() {
		errCh <- session.Run(wrapped)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case res := <-errCh:
		return res
	}
}

// Transfers data from a reader to a remote file path
func (c *Client) SendTar(ctx context.Context, reader io.Reader, remotePath string) error {
	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdin = reader

	errCh := make(chan error, 1)
	go func() {
		errCh <- session.Run(fmt.Sprintf("cat > %s && cd %s && tar -xf %s", remotePath, remotePath, remotePath))
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGTERM)
		select {
		case <-errCh:
		case <-time.After(2 * time.Second):
		}
		return ctx.Err()
	}
}

func (c *Client) RunCommandAndGetOutput(ctx context.Context, command string) (string, error) {
	session, err := c.conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	return string(output), nil
}

// Streams a remote file similar to tail -f
func (c *Client) Tail(ctx context.Context, remotePath string, stdout io.Writer) error {
	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout

	errCh := make(chan error, 1)
	go func() {
		errCh <- session.Run(fmt.Sprintf("tail -f %s", remotePath))
	}()

	select {
	case err := <-errCh:
		return err

	case <-ctx.Done():
		_ = session.Signal(ssh.SIGTERM)
		select {
		case <-errCh:
		case <-time.After(2 * time.Second):
		}
		return ctx.Err()
	}
}
