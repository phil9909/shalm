package cmd

import (
	"github.com/kramerul/shalm/pkg/shalm"
	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var applyChartArgs = shalm.ChartOptions{}

var applyCmd = &cobra.Command{
	Use:   "apply [chart]",
	Short: "apply shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exit(apply(args[0], shalm.NewK8s(), applyChartArgs.Options()))
	},
}

func apply(url string, k shalm.K8s, opts ...shalm.ChartOption) error {
	repo := shalm.NewRepo()
	thread := &starlark.Thread{Name: "main"}
	c, err := repo.Get(thread, url, opts...)
	if err != nil {
		return err
	}
	return c.Apply(thread, k)
}

func init() {
	applyChartArgs.AddFlags(applyCmd.Flags())
}
