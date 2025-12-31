package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a job",
	Long: `Run a job
(long version)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("run called")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
