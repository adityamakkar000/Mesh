package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/adityamakkar000/Mesh/internal/parse"
	"github.com/adityamakkar000/Mesh/internal/prerun"
	"github.com/adityamakkar000/Mesh/internal/ssh"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup <cluster_name>",
	Short: "Setup a cluster from mesh.yaml configuration",
	Long: `Setup a cluster by reading cluster info from ~/.config/mesh/cluster.yaml
and setup commands from ./mesh.yaml.

Example:
  mesh setup my-cluster`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]

		var code = 0
		if err := runSetup(clusterName); err != nil {
			code = 1
		}
		os.Exit(code)
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(clusterName string) error {
	error := prerun.RunPreRunSSH(clusterName, setupHost, prerun.PreRunSSHMsgs{
		HostSuccessMsg: "[%s] setup completed",
		HostErrorMsg:   "[%s] setup failed: %v",
		SuccessMsg:     "Cluster '%s' setup complete",
		ErrorMsg:       "setup failed on %d host(s)",
	})
	return error
}

func setupHost(ctx context.Context, cluster *parse.NodeConfig, mesh *parse.MeshConfig, host string) error {
	client, err := ssh.Connect(ctx, cluster.User, host, cluster.IdentityFile)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	for _, command := range mesh.Commands {
		if err := client.Exec(ctx, command, io.Discard, io.Discard); err != nil {
			return fmt.Errorf("failed to execute '%s': %w", command, err)
		}
	}

	return nil
}
