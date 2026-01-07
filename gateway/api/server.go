package api

import (
	"mock-payment-gateway/api/handlers"
	"mock-payment-gateway/api/services"
	"mock-payment-gateway/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type GatewayServer struct {
	config *config.Configuration
	server *http.Server
}

func New(config *config.Configuration) *GatewayServer {
	return &GatewayServer{
		config: config,
	}
}

func (g *GatewayServer) Serve() error {
	server := &http.Server{
		Addr:    ":" + g.config.Server.Port,
		Handler: g.routes(),
	}
	g.server = server

	log.Info().Msgf("Server started at %s", g.config.Server.Port)
	return server.ListenAndServe()
}

func (g *GatewayServer) Close() error {
	if g.server == nil {
		return nil
	}

	return g.server.Close()
}

func (g *GatewayServer) routes() http.Handler {
	route := gin.Default()

	service := services.NewPaymentService(g.config)
	handler := handlers.NewPaymentHandler(service)

	subRoute := route.Group("/api/v1/payment")
	{
		subRoute.POST("/init", handler.Process)
		subRoute.GET("/status/:id", handler.Fetch)
	}

	return route
}
