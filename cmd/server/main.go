package main

import (
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/api"
	"github.com/pedramkousari/abshar-toolbox-new/config"
)

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(1)

	cnf := config.GetCnf()
	server := api.NewServer(cnf.Server.Host, cnf.Server.Port)

	api.HandleFunc(cnf, server)
	go server.Run(wg)

	wg.Wait()
}
