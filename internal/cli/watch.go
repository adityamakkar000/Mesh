package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adityamakkar000/Mesh/internal/parse"
	"github.com/adityamakkar000/Mesh/internal/ui"
	"github.com/adityamakkar000/Mesh/internal/watch"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch <cluster_name>",
	Short: "Watch TPU metrics across cluster",
	Long: `Watch TPU utilization and memory metrics across all hosts in a cluster.
Refreshes periodically to show real-time cluster status.

Example:
  mesh watch my-cluster
  mesh watch my-cluster -n 5`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]
		interval, _ := cmd.Flags().GetInt("interval")

		if err := runWatch(clusterName, interval); err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	watchCmd.Flags().IntP("interval", "n", 2, "Refresh interval in seconds")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(clusterName string, intervalSecs int) error {
	clusters, err := parse.Clusters()
	if err != nil {
		return err
	}

	cluster, ok := clusters[clusterName]
	if !ok {
		ui.Error("cluster not found in cluster.yaml")
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		ui.Info("\nStopping watch...")
		cancel()
	}()

	ticker := time.NewTicker(time.Duration(intervalSecs) * time.Second)
	defer ticker.Stop()

	// Initial refresh
	refreshMetrics(ctx, cluster, clusterName)

	// Watch loop
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			refreshMetrics(ctx, cluster, clusterName)
		}
	}
}

func refreshMetrics(ctx context.Context, cluster parse.NodeConfig, clusterName string) {
	metrics := watch.CollectTPUMetrics(ctx, cluster)
	watch.Display(metrics, clusterName)
}
