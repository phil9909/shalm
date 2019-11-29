package cmd

import (
	"github.com/kramerul/shalm/pkg/chart/impl"

	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var deleteChartArgs = chartArgs{}

var deleteCmd = &cobra.Command{
	Use:   "delete [chart]",
	Short: "delete shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := impl.NewRepo(repoOpts()...)
		chartName := args[0]

		thread := &starlark.Thread{Name: "my thread"}
		c, err := repo.Get(thread, rootChart(), chartName, nil, deleteChartArgs.KwArgs())
		if err != nil {
			return err
		}
		return c.Delete(thread, impl.NewK8s())
	},
}

func init() {
	deleteChartArgs.AddFlags(deleteCmd.Flags())
}
