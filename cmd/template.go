package cmd

import (
	"fmt"

	"github.com/kramerul/shalm/pkg/shalm"

	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var templateChartArgs = chartArgs{}

var templateCmd = &cobra.Command{
	Use:   "template [chart]",
	Short: "template shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repo := shalm.NewRepo()
		url := args[0]

		thread := &starlark.Thread{Name: "my thread"}
		c, err := repo.Get(thread, url, rootNamespace(), nil, templateChartArgs.KwArgs())
		if err != nil {
			exit(err)
		}
		t, err := c.Template(thread)
		if err != nil {
			exit(err)
		}
		fmt.Println(t)
	},
}

func init() {
	templateChartArgs.AddFlags(templateCmd.Flags())
}
