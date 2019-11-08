package cmd

import (
	"github.wdf.sap.corp/shalm/internal/pkg/k8s"

	"github.wdf.sap.corp/shalm/internal/pkg/chart"

	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [chart]",
	Short: "delete shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var repo = chart.LocalRepo{BaseDir: repoDir}
		chartName := args[0]

		thread := &starlark.Thread{Name: "my thread"}
		c, err := chart.NewChart(thread, &repo, chartName, nil, nil)
		if err != nil {
			return err
		}
		_, err = starlark.Call(thread, c.DeleteFunction(), starlark.Tuple{&k8s.K8s{}, &chart.Release{Name: chartName, Namespace: nameSpace, Service: chartName}}, nil)
		if err != nil {
			return err
		}
		return nil
	},
}
