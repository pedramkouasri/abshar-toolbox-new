/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pedramkousari/abshar-toolbox-new/utils"
	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "update-toolbox",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func addCommands() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt Patch File",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fileSrc := args[0]

			if !utils.FileExists(fileSrc) {
				panic("File Not Found")
			}

			if err := os.Mkdir("./temp", 0755); err != nil {
				if os.IsNotExist(err) {
					panic(fmt.Errorf("create directory err: %s", err))
				}
			}

			if err := utils.DecryptFile([]byte("e10adc3949ba59abbe56e057f20f883e"), fileSrc, strings.TrimSuffix(fileSrc, ".enc")); err != nil {
				panic(fmt.Errorf("Decrypt File err:  %s", err))
			}

			if err := utils.UntarGzip(strings.TrimSuffix(fileSrc, ".enc"), "./temp"); err != nil {
				panic(fmt.Errorf("UnZip File err  %s", err))
			}
		},
	})
	// rootCmd.AddCommand(patch.PatchCmd)
	// rootCmd.AddCommand(server.ServerCmd)
}

func init() {
	cobra.OnInitialize(initConfig)
	addCommands()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {}
