package api

import (
	"net/http"
	"payment-gateway/api/service"

	"github.com/rs/zerolog/log"
)

type GatewayServer struct {
	port           string
	server         *http.Server
	serviceFactory service.ServiceFactory
}

func NewGatewayServer(port string, serviceFactory service.ServiceFactory) *GatewayServer {
	return &GatewayServer{
		port:           port,
		server:         nil,
		serviceFactory: serviceFactory,
	}
}

func (g *GatewayServer) Serve() error {
	server := &http.Server{
		Addr:    ":" + g.port,
		Handler: g.routes(),
	}
	g.server = server

	log.Info().Msgf("Server started at %s", g.port)
	return server.ListenAndServe()
}

func (g *GatewayServer) Close() error {
	if g.server == nil {
		return nil
	}

	return g.server.Close()
}
