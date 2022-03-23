package server

import (
	"github.com/utrack/clay/v3/server"
	"github.com/utrack/clay/v3/server/middlewares/mwhttp"
	"github.com/utrack/clay/v3/transport"

	"github.com/go-chi/chi"
	"google.golang.org/grpc"
)

// Option is an optional setting applied to the Server.
type Option = server.Option

// WithHTTPMiddlewares sets up HTTP middlewares to work with.
func WithHTTPMiddlewares(mws ...mwhttp.Middleware) Option {
	return server.WithHTTPMiddlewares(mws...)
}

// WithGRPCUnaryMiddlewares sets up unary middlewares for gRPC server.
func WithGRPCUnaryMiddlewares(mws ...grpc.UnaryServerInterceptor) Option {
	return server.WithGRPCUnaryMiddlewares(mws...)
}

// WithHTTPMux sets existing HTTP muxer to use instead of creating new one.
func WithHTTPMux(mux *chi.Mux) Option {
	return server.WithHTTPMux(mux)
}

func WithHTTPRouterMux(mux transport.Router) Option {
	return server.WithHTTPRouterMux(mux)
}
