package cmd

import (
	"fmt"

	"github.wdf.sap.corp/shalm/internal/pkg/chart"

	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var chartName string
var nameSpace string
var repoDir string

func init() {

	rootCmd.Flags().StringVar(&chartName, "chart", "cf", "name of the chart")
	rootCmd.Flags().StringVar(&nameSpace, "namespace", "test", "namespace")
	rootCmd.Flags().StringVar(&repoDir, "repo", "example", "directory where to find the charts")
}

var rootCmd = &cobra.Command{
	Use:   "shalm",
	Short: "Shalm brings the starlark language to helm charts",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var repo = chart.LocalRepo{BaseDir: repoDir}

		thread := &starlark.Thread{Name: "my thread"}
		c, err := chart.NewChart(thread, &repo, chartName)

		if err != nil {
			panic(err)
		}
		t, err := starlark.Call(thread, c.TemplateFunction(), starlark.Tuple{&chart.Release{Name: chartName, Namespace: nameSpace, Service: chartName}}, nil)
		if err != nil {
			panic(err)
		}
		fmt.Println(t.(starlark.String).GoString())
	},
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
