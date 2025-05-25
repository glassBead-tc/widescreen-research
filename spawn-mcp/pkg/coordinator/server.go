package coordinator

import "log"

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Serve() error {
	log.Println("Coordinator running...")
	select {}
}