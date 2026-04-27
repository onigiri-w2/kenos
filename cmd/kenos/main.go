package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:   "kenos",
		Short: "AIと一緒に働くための仕組みツール",
	}

	root.AddCommand(versionCmd())
	root.AddCommand(initCmd())
	root.AddCommand(updateCmd())
	root.AddCommand(taskCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "バージョンを表示する",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("kenos %s\n", version)
		},
	}
}
