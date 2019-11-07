package cmd

import (
	"fmt"

	"github.wdf.sap.corp/shalm/internal/pkg/chart"

	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var nameSpace string
var repoDir string

func init() {

	rootCmd.Flags().StringVar(&nameSpace, "namespace", "test", "namespace")
	rootCmd.Flags().StringVar(&repoDir, "repo", "example", "directory where to find the charts")
}

var rootCmd = &cobra.Command{
	Use:   "shalm",
	Short: "Shalm brings the starlark language to helm charts",
	Long:  ``,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var repo = chart.LocalRepo{BaseDir: repoDir}
		chartName := args[1]

		thread := &starlark.Thread{Name: "my thread"}
		c, err := chart.NewChart(thread, &repo, chartName)

		if err != nil {
			panic(err)
		}
		switch args[0] {
		case "template":
			t, err := starlark.Call(thread, c.TemplateFunction(), starlark.Tuple{&chart.Release{Name: chartName, Namespace: nameSpace, Service: chartName}}, nil)
			if err != nil {
				return err
			}
			fmt.Println(t.(starlark.String).GoString())
		case "apply":
			_, err := starlark.Call(thread, c.ApplyFunction(), starlark.Tuple{&chart.Release{Name: chartName, Namespace: nameSpace, Service: chartName}}, nil)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("Invalid command %s", args[0])
		}
		return nil
	},
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
