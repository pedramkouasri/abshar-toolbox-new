package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

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
	server.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		callStopServer()
	})

	go server.Run(stop, wg)

	go func() {
		<-rollback
		fmt.Println("RollbackStarted")
	}()

	go func() {
		<-time.After(time.Second * 20)
		rollback <- struct{}{}
	}()

	wg.Wait()

}

func callStopServer() {
	stop <- struct{}{}
}
