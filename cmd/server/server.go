package server

import (
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/api"
	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/spf13/cobra"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Run Server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		startServer(cmd)
	},
}

func startServer(cmd *cobra.Command) {
	wg := new(sync.WaitGroup)
	wg.Add(1)

	cnf := config.GetCnf()
	server := api.NewServer(cnf.Server.Host, cnf.Server.Port)

	api.HandleFunc(cnf, server)
	go server.Run(wg)

	wg.Wait()
}
