package cmd

import (
	"io"
	"os"

	"github.com/kramerul/shalm/pkg/shalm"

	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var templateChartArgs = shalm.ChartOptions{}

var templateCmd = &cobra.Command{
	Use:   "template [chart]",
	Short: "template shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exit(template(args[0], os.Stdout))
	},
}

func template(url string, writer io.Writer) error {

	thread := &starlark.Thread{Name: "main"}
	repo := shalm.NewRepo()
	c, err := repo.Get(thread, url, templateChartArgs.Options())
	if err != nil {
		return err
	}
	t, err := c.Template(thread)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(t))
	return err
}

func init() {
	templateChartArgs.AddFlags(templateCmd.Flags())
}
