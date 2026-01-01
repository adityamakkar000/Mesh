package cli

import (
	"fmt"

	"github.com/adityamakkar000/Mesh/internal/parse"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup <cluster_path> <cluster_name>",
	Short: "Setup a cluster from a YAML configuration",
	Long: `Setup a cluster by parsing the YAML configuration file and executing setup commands.

Example:
  mesh setup ./cluster.yaml my-cluster`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		clusterPath := args[0]
		clusterName := args[1]

		fmt.Printf("Setting up cluster '%s' from %s\n", clusterName, clusterPath)

		clusters, err := parse.Clusters(clusterPath)

		if err != nil {
			panic(err)
		}

		_ = clusters
		_ = clusterName
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
