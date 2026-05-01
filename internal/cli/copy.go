package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/adityamakkar000/Mesh/internal/parse"
	"github.com/adityamakkar000/Mesh/internal/prerun"
	"github.com/adityamakkar000/Mesh/internal/ssh"
	"github.com/adityamakkar000/Mesh/internal/ui"
	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy files to the cluster",
	Long: `Copy files to the cluster, mimicking a single host with pre-run commands
from ./mesh.yaml.

Example:
  mesh copy my-cluster dirNameOnCluster
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]
		remoteDir := args[1]

		ui.Info(fmt.Sprintf("Copying files to %s", remoteDir))

		code := 0
		if err := runCopy(remoteDir, clusterName); err != nil {
			ui.Error(err.Error())
			code = 1
		}
		os.Exit(code)
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
}

func runCopy(dirName, clusterName string) error {
	cluster, mesh, err := prerun.ParseConfigs(clusterName)
	if err != nil {
		return err
	}

	hostLabel := "hosts"
	if len(cluster.Hosts) == 1 {
		hostLabel = "host"
	}

	ui.Header(fmt.Sprintf(
		"Copying directory on '%s' (%d %s) to directory '%s'",
		clusterName,
		len(cluster.Hosts),
		hostLabel,
		dirName,
	))

	failures := prerun.RunOnAllHosts(
		cluster,
		mesh,
		runCopyHost(dirName),
		"[%s] Copy completed",
		"[%s] Copy failed: %v",
	)

	if failures > 0 {
		return fmt.Errorf("copy failed on %d hosts", failures)
	}

	ui.Success(fmt.Sprintf("Copy completed on %s", clusterName))
	return nil
}

func runCopyHost(dirName string) prerun.SSHCommand {
	return func(
		ctx context.Context,
		cluster *parse.NodeConfig,
		mesh *parse.MeshConfig,
		host string,
		hostID int,
	) error {

		client, err := ssh.Connect(ctx, cluster.User, host, cluster.IdentityFile)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", host, err)
		}
		defer client.Close()

		// Clean + recreate directory (safer than rm -rf *)
		cmdStr := fmt.Sprintf("rm -rf %[1]s && mkdir -p %[1]s", dirName)

		if err := client.Exec(ctx, cmdStr, io.Discard, io.Discard); err != nil {
			return fmt.Errorf("failed to execute '%s': %w", cmdStr, err)
		}

		reader := prerun.BuildTar()
		errCopy := client.SendTar(ctx, reader, dirName, mesh_file)
		if errCopy != nil {
			return fmt.Errorf("failed to send files: %w", errCopy)
		}
		return nil
	}
}