package cmd

import (
	"errors"
	"fmt"
	"os"

	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var nameSpace string = "default"

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
func rootNamespace() string {
	return nameSpace
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
