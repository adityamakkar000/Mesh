package ssh

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	conn *ssh.Client
}

func Connect(user, host, identityFile string) (*Client, error) {
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

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", host), config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Synchronous execution of a command (setup only)
func (c *Client) Exec(command string, stdout, stderr io.Writer) error {
	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr

	if err := session.Run(command); err != nil {
		return err
	}

	return nil
}

// Asynchronous execution of a command
func (c *Client) ExecDetached(command string) error {
	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// TODO(william): test this more thoroughly, and also add job id saving (like echo $$ > ~/.mesh/.jobid)
	wrapped := fmt.Sprintf("setsid %s > /dev/null 2>&1 < /dev/null &", command)
	if err := session.Start(wrapped); err != nil {
		return err
	}

	return nil
}

// Transfers data from a reader to a remote file path
func (c *Client) Copy(reader io.Reader, remotePath string, mode string) error {
	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdin = reader
	if err := session.Run(fmt.Sprintf("cat > %s && chmod %s %s", remotePath, mode, remotePath)); err != nil {
		return err
	}

	return nil
}

// Streams a remote file similar to tail -f
func (c *Client) Tail(remotePath string, stdout io.Writer) error {
	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	if err := session.Run(fmt.Sprintf("tail -f %s", remotePath)); err != nil {
		return err
	}

	return nil
}
