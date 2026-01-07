package api

import (
	"context"
	"loan-disbursement-service/api/services"
	"net/http"

	"github.com/rs/zerolog/log"
)

type DisbursementServer struct {
	port           string
	server         *http.Server
	serviceFactory *services.ServiceFactory
}

func New(port string, serviceFactory *services.ServiceFactory) *DisbursementServer {
	return &DisbursementServer{
		port:           port,
		server:         nil,
		serviceFactory: serviceFactory,
	}
}

func (d *DisbursementServer) Serve() error {
	server := &http.Server{
		Addr:    ":" + d.port,
		Handler: d.routes(),
	}
	d.server = server

	log.Info().Msgf("Disbursement server started at %s", d.port)
	return server.ListenAndServe()
}

func (d *DisbursementServer) Close(ctx context.Context) error {
	if d.server == nil {
		return nil
	}

	return d.server.Shutdown(ctx)
}
