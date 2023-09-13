/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/pedramkousari/abshar-toolbox-new/cmd/patch"
	"github.com/pedramkousari/abshar-toolbox-new/cmd/server"
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
	rootCmd.AddCommand(patch.PatchCmd)
	rootCmd.AddCommand(server.ServerCmd)
}

func init() {
	cobra.OnInitialize(initConfig)
	addCommands()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {}
