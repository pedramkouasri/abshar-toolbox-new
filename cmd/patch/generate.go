package patch

import (
	"log"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/scripts/generator"
	"github.com/spf13/cobra"
)

var PatchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Create Patch",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var generateCmd = &cobra.Command{
	Use:                   "create PATH_OF_PACKAGE.JSON",
	Short:                 "Create PATH_OF_PACKAGE.JSON",
	Long:                  ``,
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		cnf := config.GetCnf()
		pg := generator.NewPatchService(cnf)

		if err := pg.Handle(args[0]); err != nil {
			logger.Error(err)
			log.Fatalf("Generate Package Faild %v", err)
			return
		}

		log.Println("Package Build :-)")
	},
}

func init() {
	PatchCmd.AddCommand(generateCmd)
}
