package cmd

import (
	"fmt"

	repo2 "github.com/kramerul/shalm/internal/pkg/repo"

	"github.com/google/go-containerregistry/pkg/crane"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [chart]",
	Short: "push shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var repo = repo2.LocalRepo{BaseDir: repoDir}
		chartName := args[0]

		path, err := repo.Directory(chartName)
		if err != nil {
			return err
		}
		_, err = crane.Load(path)
		return fmt.Errorf("Not implemented yet")
	},
}
