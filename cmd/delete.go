package cmd

import (
	"github.com/kramerul/shalm/pkg/shalm"
	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var deleteChartArgs = shalm.ChartOptions{}

var deleteCmd = &cobra.Command{
	Use:   "delete [chart]",
	Short: "delete shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exit(delete(args[0], shalm.NewK8s(), deleteChartArgs.Options()))
	},
}

func delete(url string, k shalm.K8s, opts ...shalm.ChartOption) error {
	repo := shalm.NewRepo()
	thread := &starlark.Thread{Name: "main"}
	c, err := repo.Get(thread, url, opts...)
	if err != nil {
		return err
	}
	return c.Delete(thread, k)
}

func init() {
	deleteChartArgs.AddFlags(deleteCmd.Flags())
}
