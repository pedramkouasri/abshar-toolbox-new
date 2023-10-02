package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	http   http.Server
	mux    *http.ServeMux
	stopCh chan struct{}
}

const (
	httpAPITimeout = time.Minute * 10
	shutdwnTimeout = time.Second * 5
)

func NewServer(add string, port int) *Server {
	s := &Server{}

	s.mux = http.NewServeMux()
	s.http = http.Server{
		Addr:    fmt.Sprintf("%s:%d", add, port),
		Handler: http.TimeoutHandler(s.mux, httpAPITimeout, ""),
	}

	s.stopCh = make(chan struct{})
	return s
}

func (s *Server) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	go func() {
		if err := s.http.ListenAndServe(); err != http.ErrServerClosed {
			panic(fmt.Errorf("Could not start http server %v", err))
		}
	}()
	fmt.Printf("Listen on %s \n", s.http.Addr)

	<-s.stopCh

	ctx, cancel := context.WithTimeout(context.Background(), shutdwnTimeout)
	defer cancel()

	var err error
	if err = s.http.Shutdown(ctx); err == nil {
		fmt.Println("Http server shutdown")
		return
	}

	if err == context.DeadlineExceeded {
		fmt.Println("Shutdown timeout exceeded. closing http server")
		if err = s.http.Close(); err != nil {
			fmt.Println("Could not close http connection: %v \n", err)
		}
		return
	}

	fmt.Println("Could not close shutdown http server: %v \n", err)
}

func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *Server) Stop() {
	s.stopCh <- struct{}{}
}
