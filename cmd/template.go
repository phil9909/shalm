package cmd

import (
	"fmt"

	"go.starlark.net/starlark"

	"github.com/kramerul/shalm/internal/pkg/chart/api"
	"github.com/kramerul/shalm/internal/pkg/chart/impl"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template [chart]",
	Short: "template shalm chart",
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
		t, err := c.Template(thread, &api.Release{Name: chartName, Namespace: nameSpace, Service: chartName})
		fmt.Println(t)
		return nil
	},
}
