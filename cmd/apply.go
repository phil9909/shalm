package cmd

import (
	"github.com/kramerul/shalm/internal/pkg/k8s"
	"github.com/kramerul/shalm/internal/pkg/repo"

	"github.com/kramerul/shalm/internal/pkg/chart"

	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [chart]",
	Short: "apply shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var repo = repo.LocalRepo{BaseDir: repoDir}
		chartName := args[0]
		return apply(&repo, chartName, k8s.New(), &chart.Release{Name: chartName, Namespace: nameSpace, Service: chartName})
	},
}

func apply(repo repo.Repo, chartName string, k k8s.K8s, release *chart.Release) error {
	thread := &starlark.Thread{Name: "my thread"}
	c, err := chart.NewChart(thread, repo, chartName, nil, nil)
	if err != nil {
		return err
	}
	_, err = starlark.Call(thread, c.ApplyFunction(), starlark.Tuple{k, release}, nil)
	if err != nil {
		return err
	}
	return nil
}
