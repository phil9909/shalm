package cmd

import (
	"os"

	"github.com/kramerul/shalm/pkg/shalm"

	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var packageCmd = &cobra.Command{
	Use:   "package [chart]",
	Short: "package shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exit(pkg(args[0], rootNamespace()))
	},
}

func pkg(url string, namespace string) error {
	repo := shalm.NewRepo()

	thread := &starlark.Thread{Name: "main"}
	c, err := repo.Get(thread, url, rootNamespace(), nil, nil)
	if err != nil {
		return err
	}
	out, err := os.Create(c.GetName() + "-" + c.GetVersion().String() + ".tgz")
	if err != nil {
		return err
	}
	defer out.Close()
	return c.Package(out)
}
