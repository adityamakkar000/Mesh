package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/adityamakkar000/Mesh/internal/parse"
	"github.com/adityamakkar000/Mesh/internal/ssh"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch <cluster_name>",
	Short: "Watch the tpu-info for a given cluster",
	Long: `Watch the tpu-info for a given cluster"

Example:
  mesh watch my-cluster`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]

		var code = 0
		if err := runWatch(clusterName); err != nil {
			code = 1
		}
		os.Exit(code)
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}

var get_tpu_info_cmd = `
# adapated from: https://github.com/rdyro/nvtop-with-tpu/commit/f5b3978c3b115eb0c125a0a44c74d70c420a75c3
try:
	from tpu_info import device, metrics
except Exception as _:
	raise RuntimeError("tpu_info missing")

try:
	chip_type, count = device.get_local_chips()
	print(chip_type)
	print(count)
except Exception as _:
	raise RuntimeError("Failed to get TPU info, check if you are runnning a TPU framework on all hosts")
`

var get_tpu_metrics_cmd = `
# adapated from: https://github.com/rdyro/nvtop-with-tpu/commit/f5b3978c3b115eb0c125a0a44c74d70c420a75c3
try:
	from tpu_info import device, metrics
except Exception as _:
	raise RuntimeError("tpu_info missing")

try:
	chip_type, count = device.get_local_chips()
	chips_usage = metrics.get_chip_usage(chip_type)

	for chip_usage in chips_usage:
		print(
			f"{chip_type.value.name}-{chip_usage.device_id:d} "
			f"{chip_usage.memory_usage:d} "
			f"{chip_usage.total_memory:d} "
			f"{chip_usage.duty_cycle_pct:.4f} "
			,
			flush=True,
		)


except Exception as _:
	raise RuntimeError("Failed to get TPU info, check if you are runnning a TPU framework on all hosts")
`

func runWatch(clusterName string) error {

	_, err := get_info(context.Background(), &parse.NodeConfig{}, &parse.MeshConfig{}, "")
	if err != nil {
		return err
	}
	return nil

}

type TPUInfo struct {
	ChipType string
	Count    int
}

type TPUINFO struct {
	ChipType string
	Count int 
}


func get_info(
	ctx context.Context,
	cluster *parse.NodeConfig,
	mesh *parse.MeshConfig,
	host string,
) (TPUInfo, error) {	
	client, err := ssh.Connect(ctx, cluster.User, host, cluster.IdentityFile)
	if err != nil {
		return TPUInfo{}, fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()
	cmd := fmt.Sprintf(`mkdir -p watch2 && rm -rf watch2/* && cd watch2 &&
pip install --user uv > /dev/null 2>&1 &&
export PATH="$HOME/.local/bin:$PATH" &&
uv venv --clear --python 3.12 > /dev/null 2>&1 &&
source .venv/bin/activate &&
uv pip install tpu-info > /dev/null 2>&1 &&
uv pip install 'jax[tpu]' > /dev/null 2>&1 &&
python << 'EOF'
%s
EOFk`, get_tpu_info_cmd,
)
	output, err := client.RunCommandAndGetOutput(ctx, cmd)
	if err != nil {
		return TPUInfo{}, fmt.Errorf("failed to run tpu info command: %w", err)
	}
	lines := strings.Fields(output)

	if len(lines) != 4 {
		return TPUInfo{}, fmt.Errorf("invalid tpu info output")
	}
	chipType := lines[0] + " " + lines[1]
	
	count, err := strconv.Atoi(strings.TrimSpace(lines[3]))
	if err != nil {
		return TPUInfo{}, fmt.Errorf("failed to parse chip count: %w", err)
	}

	return TPUInfo{
		ChipType: chipType,
		Count:    count,
	}, nil
}
