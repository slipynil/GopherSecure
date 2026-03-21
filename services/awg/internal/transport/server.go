package transport

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type handlers interface {
	AddPeer(w http.ResponseWriter, r *http.Request)
	DeletePeer(w http.ResponseWriter, r *http.Request)
	SendConfFile(w http.ResponseWriter, r *http.Request)
	GetKeys(w http.ResponseWriter, r *http.Request)
}

type server struct {
	handlers handlers
}

func New(h handlers) *server {
	return &server{
		handlers: h,
	}
}

func (s *server) Start(endpoint string) {
	r := mux.NewRouter()
	r.HandleFunc("/peers", s.handlers.DeletePeer).Methods("DELETE")
	r.HandleFunc("/peers", s.handlers.AddPeer).Methods("POST")
	r.HandleFunc("/peers/{id}/config", s.handlers.SendConfFile).Methods("GET")
	r.HandleFunc("/peers/{id}/keys", s.handlers.GetKeys).Methods("GET")

	fmt.Printf("HTTP started on %s\n", endpoint)
	http.ListenAndServe(endpoint, r)
}
