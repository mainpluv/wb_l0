package delivery

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	server *http.Server
}

func NewServer(router *mux.Router) *Server {
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	return &Server{server: server}
}

func (s *Server) RunServer() error {
	err := s.server.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}
