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

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a job",
	Long: `Run a command on the cluster, mimicking a single host with pre-run commands
from ./mesh.yaml.

Example:
  mesh run my-cluster python main.py --lr 1e-3

Note it passes in the rank of the process so host x will run on each cluster 'RANK=x python main.py --lr 1e-3'
Be sure to initalize JAX with induvial ranks using ENV variables to display proper logs from process 0
`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]
		command := strings.Join(args[1:], " ")
		ui.Info(fmt.Sprintf("Running command: %s", command))
		var code = 0
		if err := run(clusterName, command); err != nil {
			code = 1
		}
		os.Exit(code)

	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func run(clusterName, command string) error {

	error := prerun.RunPreRunSSH(clusterName, runHost(command), prerun.PreRunSSHMsgs{
		HostSuccessMsg: "[%s] Run completed",
		HostErrorMsg:   "[%s] Run failed: %v",
		SuccessMsg:     "Cluster '%s' run complete",
		ErrorMsg:       "run failed on %d host(s)",
	})
	return error
}

func runHost(command string) prerun.SSHCommand {

	var remote_dir = "job"
	var mesh_file = "mesh.tar"
	var log_file = "output.log"
	return func(ctx context.Context, cluster *parse.NodeConfig, mesh *parse.MeshConfig, host string) error {

		client, err := ssh.Connect(ctx, cluster.User, host, cluster.IdentityFile)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer client.Close()

		if err := client.Exec(ctx, fmt.Sprintf("mkdir -p %s && rm -rf %s/*", remote_dir, remote_dir), io.Discard, io.Discard); err != nil {
			return fmt.Errorf("failed to execute '%s': %w", command, err)
		}

		reader := prerun.BuildTar()
		errCopy := client.SendTar(ctx, reader, remote_dir, mesh_file)

		if errCopy != nil {
			return fmt.Errorf("failed to send files: %w", errCopy)
		}

		ui.Success(fmt.Sprintf("[%s] Directory built and copied", host))

		for _, command := range mesh.Prerun {
			if err := client.Exec(ctx, command, io.Discard, io.Discard); err != nil {
				return fmt.Errorf("failed to execute '%s': %w", command, err)
			}
		}

		ui.Success(fmt.Sprintf("[%s] Pre-run commands executed", host))


		if err := client.ExecDetached(ctx, command, remote_dir, log_file); err != nil {
			return fmt.Errorf("failed to execute command: %w", err)
		}

		ui.Success(fmt.Sprintf("[%s] Command started", host))

		prefixWriter := ui.NewPrefixWriter(fmt.Sprintf("[%s] ", host), os.Stdout)
		if err := client.Tail(ctx, remote_dir, log_file, prefixWriter); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("failed to tail logs: %w", err)
		}

		if err := client.Exec(ctx, "rm -rf job", io.Discard, io.Discard); err != nil {
			return fmt.Errorf("failed to execute cleanup: %w", err)
		}

		ui.Success(fmt.Sprintf("[%s] Cleaned up job directory", host))

		return nil

	}
}
