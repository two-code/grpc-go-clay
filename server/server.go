package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/utrack/clay/v3/server/log"
	"github.com/utrack/clay/v3/transport"
)

// Server is a transport server.
type Server struct {
	opts *serverOpts
}

// NewServer creates a Server.
// Pass additional Options to mutate its behaviour.
//
func NewServer(opts ...Option) *Server {
	serverOpts := defaultServerOpts()
	for _, opt := range opts {
		opt(serverOpts)
	}
	return &Server{opts: serverOpts}
}

// Run starts processing requests to the service.
// It blocks indefinitely, run asynchronously to do anything after that.
//
// Deprecated: Call Serve method instead.
func (s *Server) Run(svc transport.Service) error {
	return fmt.Errorf("not supported any more; use Serve method")
}

func initializeHttpRouter(desc transport.ServiceDesc, opts *serverOpts) chi.Router {
	router := chi.NewMux()

	if len(opts.HTTPMiddlewares) > 0 {
		router.Use(opts.HTTPMiddlewares...)
	}

	router.Mount("/", opts.HTTPMux)

	router.HandleFunc(
		"/swagger.json",
		func(w http.ResponseWriter, req *http.Request) {
			_, _ = io.Copy(w, bytes.NewReader(desc.SwaggerDef()))
		})

	return router
}

func (s *Server) Serve(
	stopCtx context.Context,
	desc transport.ServiceDesc,
	httpListener net.Listener,
	logCb func(ctx context.Context, lvl log.Level, msg string),
) (err error) {
	httpRouter := initializeHttpRouter(desc, s.opts)

	// apply gRPC interceptor
	//
	if configurableServiceDesc, ok := desc.(transport.ConfigurableServiceDesc); ok {
		configurableServiceDesc.Apply(transport.WithUnaryInterceptor(s.opts.GRPCUnaryInterceptor))
	}

	// register HTTP.
	//
	desc.RegisterHTTP(httpRouter)

	httpServer := &http.Server{
		Handler: httpRouter,
	}

	shutdownCtx, shutdownCtxCancel := context.WithCancel(stopCtx)

	shutdownChErr := make(chan error, 1)
	defer close(shutdownChErr)

	go func() {
		<-shutdownCtx.Done()

		if logCb != nil {
			logCb(stopCtx, log.LevelInfo, fmt.Sprintf("attempt to stop HTTP-server gracefully on %s ...", httpListener.Addr()))
		}

		if s.opts.HTTPShutdownWaitTimeout == -1 {
			// infinite timeout.
			//
			shutdownChErr <- httpServer.Shutdown(context.Background())

			return
		}

		waitCtx, waitCtxCancel := context.WithTimeout(context.Background(), s.opts.HTTPShutdownWaitTimeout)
		defer waitCtxCancel()

		shutdownChErr <- httpServer.Shutdown(waitCtx)
	}()

	if logCb != nil {
		logCb(stopCtx, log.LevelInfo, fmt.Sprintf("start serving HTTP-server on %s ...", httpListener.Addr()))
	}

	serveErr := httpServer.Serve(httpListener)
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		err = fmt.Errorf("error while serving HTTP-server on %s; err: %w", httpListener.Addr(), serveErr)
	}

	shutdownCtxCancel()

	shutdownErr := <-shutdownChErr
	if shutdownErr != nil {
		if err == nil {
			err = fmt.Errorf("error while shutting down HTTP-server on %s; err: %w", httpListener.Addr(), shutdownErr)
		} else {
			err = fmt.Errorf("%s; %w", err, shutdownErr)
		}
	}

	return
}
