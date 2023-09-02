package main

import (
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/api"
	"github.com/pedramkousari/abshar-toolbox-new/config"
)

var stop chan struct{} = make(chan struct{})
var rollback chan struct{} = make(chan struct{})

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(1)

	cnf := config.GetCnf()
	server := api.NewServer(cnf.Server.Host, cnf.Server.Port)

	api.HandleFunc(server)
	go server.Run(wg)

	wg.Wait()
}
