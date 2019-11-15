package cmd

import (
	"github.com/kramerul/shalm/internal/pkg/chart/impl"
	"github.com/spf13/cobra"
	"go.starlark.net/starlark"
)

var username string
var password string

var pushCmd = &cobra.Command{
	Use:   "push [chart] [tag]",
	Short: "push shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		thread := &starlark.Thread{Name: "my thread"}
		localRepo := &impl.LocalRepo{BaseDir: repoDir}
		repo := impl.NewOciRepo(func(repo string) (string, string, error) {
			// return "_json_key", os.Getenv("GCR_ADMIN_CREDENTIALS"), nil
			return username, password, nil
		})
		chart, err := localRepo.Get(thread, args[0], nil, nil)
		if err != nil {
			return err
		}
		return repo.Push(chart, args[1])
	},
}

func init() {
	pushCmd.Flags().StringVarP(&username, "user", "u", "", "user for docker login")
	pushCmd.Flags().StringVarP(&password, "password", "p", "", "password for docker login")
}
