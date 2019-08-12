package server

import (
	"eth2wtf-server/client"
	"eth2wtf-server/hub"
	"net/http"
)

type Server struct {
	clientHub *hub.Hub
	world *World
}

func NewServer() *Server {
	return &Server{
		clientHub: hub.NewHub(),
		world: NewWorld(),
	}
}

func (s *Server) Run() {
	s.clientHub.Run()
}

func (s *Server) ServeWs(w http.ResponseWriter, r *http.Request) {
	s.clientHub.ServeWs(w, r, s.NewClientHandler)
}

func (s *Server) NewClientHandler(send chan<- []byte) client.ClientHandler {
	return NewClientHandler(s.world, send)
}
