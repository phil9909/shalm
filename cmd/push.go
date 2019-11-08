package cmd

import (
	"fmt"

	"github.com/kramerul/shalm/internal/pkg/chart"

	"github.com/google/go-containerregistry/pkg/crane"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [chart]",
	Short: "push shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var repo = chart.LocalRepo{BaseDir: repoDir}
		chartName := args[0]

		path, err := repo.Directory(chartName)
		if err != nil {
			return err
		}
		_, err = crane.Load(path)
		return fmt.Errorf("Not implemented yet")
	},
}
