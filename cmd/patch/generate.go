package patch

import (
	"fmt"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/scripts/generator"
	"github.com/spf13/cobra"
)

var PatchCmd = &cobra.Command{
	Use:                   "create PATH_OF_PACKAGE.JSON",
	Short:                 "Create PATH_OF_PACKAGE.JSON",
	Long:                  ``,
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		cnf := config.GetCnf()
		pg := generator.NewPatchService(cnf)

		if err := pg.Handle(args[0]); err != nil {
			fmt.Errorf("Generate Package Faild %v", err)
		}
	},
}
