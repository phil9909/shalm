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
		chartName := args[0]
		repo := impl.NewRepo(authOpts()...)
		return apply(repo, rootChart(), chartName, impl.NewK8s())
	},
}

func apply(repo api.Repo, parent api.Chart, chartName string, k api.K8s) error {
	thread := &starlark.Thread{Name: "my thread"}
	c, err := repo.Get(thread, parent, chartName, nil, nil)
	if err != nil {
		return err
	}
	return c.Apply(thread, k)
}
