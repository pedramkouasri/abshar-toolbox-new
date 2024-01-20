package server

import (
	"fmt"
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/api"
	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
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
	ip, err := utils.GetInterfaceIpv4Addr("docker0")
	if err != nil {
		fmt.Errorf("err is :%s", err)
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)

	cnf := config.GetCnf()
	server := api.NewServer(ip, 9990)

	api.HandleFunc(cnf, server)
	go server.Run(wg)

	wg.Wait()
}
