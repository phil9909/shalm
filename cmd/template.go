package cmd

import (
	"fmt"

	repo2 "github.com/kramerul/shalm/internal/pkg/repo"

	"github.com/kramerul/shalm/internal/pkg/chart"

	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template [chart]",
	Short: "template shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var repo = repo2.LocalRepo{BaseDir: repoDir}
		chartName := args[0]

		thread := &starlark.Thread{Name: "my thread"}
		c, err := chart.NewChart(thread, &repo, chartName, nil, nil)
		if err != nil {
			return err
		}
		t, err := starlark.Call(thread, c.TemplateFunction(), starlark.Tuple{&chart.Release{Name: chartName, Namespace: nameSpace, Service: chartName}}, nil)
		if err != nil {
			return err
		}
		fmt.Println(t.(starlark.String).GoString())
		return nil
	},
}
