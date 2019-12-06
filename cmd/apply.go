package cmd

import (
	"github.com/kramerul/shalm/pkg/chart"
	"github.com/kramerul/shalm/pkg/chart/impl"
	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var applyChartArgs = chartArgs{}

var applyCmd = &cobra.Command{
	Use:   "apply [chart]",
	Short: "apply shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		repo := impl.NewRepo()
		exit(apply(repo, url, rootNamespace(), impl.NewK8s()))
	},
}

func apply(repo chart.Repo, url string, namespace string, k chart.K8s) error {
	thread := &starlark.Thread{Name: "my thread"}
	c, err := repo.Get(thread, url, namespace, nil, applyChartArgs.KwArgs())
	if err != nil {
		return err
	}
	return c.Apply(thread, k)
}

func init() {
	applyChartArgs.AddFlags(applyCmd.Flags())
}
