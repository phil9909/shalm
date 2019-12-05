package cmd

import (
	"errors"
	"fmt"
	"os"

	"go.starlark.net/starlark"

	"github.com/kramerul/shalm/pkg/chart"
	"github.com/kramerul/shalm/pkg/chart/impl"
	"github.com/spf13/cobra"
)

var nameSpace string = "default"
var username string
var password string

func init() {

	rootCmd.PersistentFlags().StringVarP(&nameSpace, "namespace", "n", "default", "namespace")
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(packageCmd)
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

func repoOpts() []impl.RepoOpts {
	if username != "" {
		return []impl.RepoOpts{impl.WithAuthCreds(func(repo string) (string, string, error) {
			// return "_json_key", os.Getenv("GCR_ADMIN_CREDENTIALS"), nil
			return username, password, nil
		})}
	}
	return nil
}

func rootChart() chart.Chart {
	chart, err := impl.NewRootChart(nameSpace)
	if err != nil {
		panic(err)
	}
	return chart
}

func unwrapEvalError(err error) error {
	if err == nil {
		return nil
	}
	evalError, ok := err.(*starlark.EvalError)
	if ok {
		return errors.New(evalError.Backtrace())
	}
	return err
}

func exit(err error) {
	if err != nil {
		fmt.Println(unwrapEvalError(err).Error())
		os.Exit(1)
	}
	os.Exit(0)
}
