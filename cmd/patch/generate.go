package patch

import (
	"fmt"

	"github.com/spf13/cobra"
)

var PatchCmd = &cobra.Command{
	Use:                   "create PATH_OF_PACKAGE.JSON",
	Short:                 "Create PATH_OF_PACKAGE.JSON",
	Long:                  ``,
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Patch Generator")
	},
}
