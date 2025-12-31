package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "Mesh",
	Short: "Insert description for mesh here",
	Long: `Mesh is a CLI written in Go that [insert purpose].
This application is a tool to [insert more].`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
