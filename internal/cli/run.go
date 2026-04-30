package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"strings"

	"github.com/adityamakkar000/Mesh/internal/parse"
	"github.com/adityamakkar000/Mesh/internal/prerun"
	"github.com/adityamakkar000/Mesh/internal/ssh"
	"github.com/adityamakkar000/Mesh/internal/ui"
	"github.com/spf13/cobra"
)

var remote_dir = "job"
var mesh_file = "mesh.tar"
var log_file = "output.log"

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a job",
	Long: `Run a command on the cluster, mimicking a single host with pre-run commands
from mesh.yaml. Use --dir to read mesh.yaml from another directory; files uploaded to
hosts are always taken from the directory you run mesh from (the process working directory).

Example:
  mesh run my-cluster python main.py --lr 1e-3
  mesh run --dir /path/to/configs my-cluster python main.py

Note it passes in the rank of the process so host x will run on each cluster 'RANK=x python main.py --lr 1e-3'
Be sure to initialize JAX with individual ranks using ENV variables to display proper logs from process 0
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]
		command := strings.Join(args[1:], " ")
		if command == "" {
			command = "true"
			ui.Warn(fmt.Sprintf("No command provided for cluster '%s'; running default no-op command '%s'.", clusterName, command))
		}
		ui.Info(fmt.Sprintf("Running command: %s", command))
		meshYAMLDir, err := resolveMeshYAMLDir()
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		var code = 0
		if err := run(clusterName, command, meshYAMLDir); err != nil {
			code = 1
		}
		os.Exit(code)

	},
}

func init() {
	runCmd.Flags().StringVar(&meshYAMLDirFlag, "dir", "", "directory containing mesh.yaml to use (default: current directory); upload tree is still the cwd where you invoke mesh")
	rootCmd.AddCommand(runCmd)
}

func run(clusterName, command, meshYAMLDir string) error {
	cluster, mesh, err := prerun.ParseConfigs(clusterName, meshYAMLDir)
	if err != nil {
		return err
	}
	var hostLabel string
	if len(cluster.Hosts) == 1 {
		hostLabel = "host"
	} else {
		hostLabel = "hosts"
	}
	ui.Header(fmt.Sprintf("Launching run on '%s' (%d %s) with command '%s'", clusterName, len(cluster.Hosts), hostLabel, command))

	failures := prerun.RunOnAllHosts(cluster, mesh, runHost(meshYAMLDir, command),
		"[%s] Run completed",
		"[%s] Run failed: %v",
	)
	if failures > 0 {
		return fmt.Errorf("Run failed on %d hosts", failures)
	}

	cleanupFailures := prerun.RunOnAllHosts(cluster, mesh, runCleanupHost,
		"[%s] Cleanup completed",
		"[%s] Cleanup failed: %v",
	)
	if cleanupFailures > 0 {
		return fmt.Errorf("Cleanup failed on %d hosts", cleanupFailures)
	}

	ui.Success(fmt.Sprintf("Run completed on %s", clusterName))
	return nil

}
func runCleanupHost(ctx context.Context, cluster *parse.NodeConfig, mesh *parse.MeshConfig, host string, host_id int) error {
	client, err := ssh.Connect(ctx, cluster.User, host, cluster.IdentityFile)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// in theory this errors if the trainin job actually finished not sure why
	// so just don't do error handling here
	_ = client.Exec(ctx, fmt.Sprintf("rm -rf %s && pkill -9 python", remote_dir), io.Discard, io.Discard)
	return nil
}

func runHost(meshYAMLDir, command string) prerun.SSHCommand {
	return func(ctx context.Context, cluster *parse.NodeConfig, mesh *parse.MeshConfig, host string, host_id int) error {

		client, err := ssh.Connect(ctx, cluster.User, host, cluster.IdentityFile)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer client.Close()

		if err := client.Exec(ctx, fmt.Sprintf("rm -rf %s && mkdir -p %s", remote_dir, remote_dir), io.Discard, io.Discard); err != nil {
			return fmt.Errorf("failed to execute '%s': %w", command, err)
		}

		reader := prerun.BuildTar(meshYAMLDir)
		errCopy := client.SendTar(ctx, reader, remote_dir, mesh_file)

		if errCopy != nil {
			return fmt.Errorf("failed to send files: %w", errCopy)
		}

		ui.Success(fmt.Sprintf("[%s] Directory built and copied", host))

		var prerun_final_command []string

		for _, command := range mesh.Prerun {
			prerun_final_command = append(prerun_final_command, command)
		}

		if err := client.ExecDetached(ctx, prerun_final_command, command, remote_dir, log_file, host_id); err != nil {
			return fmt.Errorf("failed to execute command: %w", err)
		}

		ui.Success(fmt.Sprintf("[%s] Command launched", host))

		prefixWriter := ui.NewPrefixWriter(fmt.Sprintf("[%s] ", host), os.Stdout, host_id)
		if err := client.Tail(ctx, remote_dir, log_file, prefixWriter); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("failed to tail logs: %w", err)
		}

		return nil

	}
}
