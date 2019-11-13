package cmd

import (
	"github.com/kramerul/shalm/internal/pkg/chart/api"
	"github.com/kramerul/shalm/internal/pkg/chart/impl"

	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [chart]",
	Short: "apply shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var repo = impl.LocalRepo{BaseDir: repoDir}
		chartName := args[0]
		return apply(&repo, chartName, impl.NewK8s(), &api.Release{Name: chartName, Namespace: nameSpace, Service: chartName})
	},
}

func apply(repo api.Repo, chartName string, k starlark.Value, release *api.Release) error {
	thread := &starlark.Thread{Name: "my thread"}
	c, err := repo.Get(thread, chartName, nil, nil)
	if err != nil {
		return err
	}
	_, err = starlark.Call(thread, c.ApplyFunction(), starlark.Tuple{k, impl.NewReleaseValue(release)}, nil)
	if err != nil {
		return err
	}
	return nil
}
