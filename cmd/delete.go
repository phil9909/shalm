package cmd

import (
	"github.com/kramerul/shalm/internal/pkg/chart/api"
	"github.com/kramerul/shalm/internal/pkg/chart/impl"

	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [chart]",
	Short: "delete shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var repo = impl.LocalRepo{BaseDir: repoDir}
		chartName := args[0]

		thread := &starlark.Thread{Name: "my thread"}
		c, err := repo.Get(thread, chartName, nil, nil)
		if err != nil {
			return err
		}
		_, err = starlark.Call(thread, c.DeleteFunction(), starlark.Tuple{impl.NewK8s(), impl.NewReleaseValue(&api.Release{Name: chartName, Namespace: nameSpace, Service: chartName})}, nil)
		if err != nil {
			return err
		}
		return nil
	},
}
