package api

import (
	"net/http"
)

func HandleFunc(server *Server) {
	server.HandleFunc("/ping", pingHandle)
	server.HandleFunc("/patch", patchHandle)
	server.HandleFunc("/state", stateHandle)
}

func pingHandle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func patchHandle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}
func stateHandle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}
