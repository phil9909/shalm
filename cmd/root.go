package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var nameSpace string
var repoDir string

func init() {

	rootCmd.PersistentFlags().StringVarP(&nameSpace, "namespace", "n", "default", "namespace")
	rootCmd.PersistentFlags().StringVarP(&repoDir, "repo", "r", "example", "directory where to find the charts")
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(pushCmd)
}

var rootCmd = &cobra.Command{
	Use:   "shalm",
	Short: "Shalm brings the starlark language to helm charts",
	Long:  ``,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Invalid command %s", args[0])
	},
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
