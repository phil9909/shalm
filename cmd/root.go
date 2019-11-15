package cmd

import (
	"fmt"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/spf13/cobra"
)

var nameSpace string
var repoDir string
var username string
var password string

func init() {

	rootCmd.PersistentFlags().StringVarP(&nameSpace, "namespace", "n", "default", "namespace")
	rootCmd.PersistentFlags().StringVarP(&repoDir, "repo", "r", "example", "directory where to find the charts")
	pushCmd.PersistentFlags().StringVarP(&username, "user", "u", "", "user for docker login")
	pushCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "password for docker login")
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

func authOpts() []docker.AuthorizerOpt {
	if username != "" {
		return []docker.AuthorizerOpt{docker.WithAuthCreds(func(repo string) (string, string, error) {
			// return "_json_key", os.Getenv("GCR_ADMIN_CREDENTIALS"), nil
			return username, password, nil
		})}
	}
	return nil
}
