package cmd

import (
	"github.com/kramerul/shalm/pkg/chart/impl"
	"github.com/spf13/cobra"
	"go.starlark.net/starlark"
)

var pushCmd = &cobra.Command{
	Use:   "push [chart] [tag]",
	Short: "push shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		thread := &starlark.Thread{Name: "my thread"}

		repo := impl.NewRepo(repoOpts()...)
		chart, err := repo.Get(thread, rootChart(), args[0], nil, nil)
		if err != nil {
			return err
		}
		return repo.Push(chart, args[1])
	},
}

func init() {
}
