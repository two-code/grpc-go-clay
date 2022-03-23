package server

import (
	"net/http"
	"time"

	"github.com/utrack/clay/v3/server/middlewares/mwhttp"
	"github.com/utrack/clay/v3/transport"

	"github.com/go-chi/chi"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// Option is an optional setting applied to the Server.
type Option func(*serverOpts)

type serverOpts struct {
	HTTPMux                 transport.Router
	GRPCUnaryInterceptor    grpc.UnaryServerInterceptor
	HTTPMiddlewares         []func(http.Handler) http.Handler
	HTTPShutdownWaitTimeout time.Duration
}

func defaultServerOpts() *serverOpts {
	return &serverOpts{
		HTTPMux: chi.NewMux(),
	}
}

// WithHTTPMiddlewares sets up HTTP middlewares to work with.
func WithHTTPMiddlewares(mws ...mwhttp.Middleware) Option {
	mwGeneric := make([]func(http.Handler) http.Handler, 0, len(mws))
	for _, mw := range mws {
		mwGeneric = append(mwGeneric, mw)
	}
	return func(o *serverOpts) {
		o.HTTPMiddlewares = mwGeneric
	}
}

func WithHTTPShutdownWaitTimeout(timeout time.Duration) Option {
	return func(o *serverOpts) {
		o.HTTPShutdownWaitTimeout = timeout
	}
}

// WithGRPCUnaryMiddlewares sets up unary middlewares for gRPC server.
func WithGRPCUnaryMiddlewares(mws ...grpc.UnaryServerInterceptor) Option {
	mw := grpc_middleware.ChainUnaryServer(mws...)
	return func(o *serverOpts) {
		o.GRPCUnaryInterceptor = mw
	}
}

// WithHTTPMux sets existing HTTP muxer to use instead of creating new one.
func WithHTTPMux(mux *chi.Mux) Option {
	return func(o *serverOpts) {
		o.HTTPMux = mux
	}
}

func WithHTTPRouterMux(mux transport.Router) Option {
	return func(o *serverOpts) {
		o.HTTPMux = mux
	}
}
