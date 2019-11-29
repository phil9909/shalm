package cmd

import (
	"os"

	"github.com/kramerul/shalm/pkg/chart/impl"

	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var packageCmd = &cobra.Command{
	Use:   "package [chart]",
	Short: "package shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := impl.NewRepo(repoOpts()...)
		chartName := args[0]

		thread := &starlark.Thread{Name: "my thread"}
		c, err := repo.Get(thread, rootChart(), chartName, nil, nil)
		if err != nil {
			return err
		}
		out, err := os.Create(c.GetName() + "-" + c.GetVersion().String() + ".tgz")
		if err != nil {
			return err
		}
		defer out.Close()
		return c.Package(out)
	},
}
