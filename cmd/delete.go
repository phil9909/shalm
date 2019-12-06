package cmd

import (
	"github.com/kramerul/shalm/pkg/shalm"
	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var deleteChartArgs = chartArgs{}

var deleteCmd = &cobra.Command{
	Use:   "delete [chart]",
	Short: "delete shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exit(delete(args[0], rootNamespace(), shalm.NewK8s()))
	},
}

func delete(url string, namespace string, k shalm.K8s) error {
	repo := shalm.NewRepo()
	thread := &starlark.Thread{Name: "main"}
	c, err := repo.Get(thread, url, namespace, nil, applyChartArgs.KwArgs())
	if err != nil {
		return err
	}
	return c.Delete(thread, k)
}

func init() {
	deleteChartArgs.AddFlags(deleteCmd.Flags())
}
