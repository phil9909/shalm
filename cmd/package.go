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
	Run: func(cmd *cobra.Command, args []string) {
		repo := impl.NewRepo()
		url := args[0]

		thread := &starlark.Thread{Name: "my thread"}
		c, err := repo.Get(thread, url, rootNamespace(), nil, nil)
		if err != nil {
			exit(err)
		}
		out, err := os.Create(c.GetName() + "-" + c.GetVersion().String() + ".tgz")
		if err != nil {
			exit(err)
		}
		defer out.Close()
		exit(c.Package(out))
	},
}
