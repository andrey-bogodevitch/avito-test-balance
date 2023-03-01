package api

import "net/http"

type Server struct {
	router *http.ServeMux
	http   *http.Server
}

func NewServer(port string, uh *UserHandler) *Server {
	r := http.NewServeMux()
	r.HandleFunc("/userbalance", uh.GetUserBalance)
	r.HandleFunc("/increase", uh.IncreaseBalance)
	r.HandleFunc("/decrease", uh.DecreaseBalance)
	r.HandleFunc("/transfer", uh.TransferMoney)
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	server := &Server{
		router: r,
		http:   httpServer,
	}
	return server
}

func (s *Server) Run() error {
	return s.http.ListenAndServe()
}
