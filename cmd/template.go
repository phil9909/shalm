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
		repo := impl.NewRepo(authOpts()...)
		chartName := args[0]

		thread := &starlark.Thread{Name: "my thread"}
		c, err := repo.Get(thread, nil, chartName, nil, nil)
		if err != nil {
			return err
		}
		t, err := c.Template(thread, &api.InstallOpts{Namespace: nameSpace})
		fmt.Println(t)
		return nil
	},
}
