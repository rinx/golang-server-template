package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/kpango/glg"
	"github.com/kpango/golang-server-template/config"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Server represents server behavior
type Server interface {
	ListenAndServe(context.Context) chan []error
}

type server struct {
	// api server
	srv *http.Server

	// Health Check server
	hcsrv *http.Server

	// grpc server
	grpcsrv *grpc.Server

	// grpc web server
	gwebsrv *http.Server

	cfg config.Server

	// ProbeWaitTime
	pwt time.Duration

	// ShutdownDuration
	sddur time.Duration
}

const (
	// ContentType represents a HTTP header name "Content-Type"
	ContentType = "Content-Type"

	// TextPlain represents a HTTP content type "text/plain"
	TextPlain = "text/plain"

	// CharsetUTF8 represents a UTF-8 charset for HTTP response "charset=UTF-8"
	CharsetUTF8 = "charset=UTF-8"
)

var (
	// ErrContextClosed represents a error that the context is closed
	ErrContextClosed = errors.New("context Closed")
)

// NewServer returns a Server interface, which includes api server and health check server structs.
// The api server is a http.Server instance, which the port number is read from "config.Server.Port"
// , and set the handler as this function argument "handler".
//
// The health check server is a http.Server instance, which the port number is read from "config.Server.HealthzPort"
// , and the handler is as follow - Handle HTTP GET request and always return HTTP Status OK (200) response.
func NewServer(cfg config.Server, h http.Handler, g *grpc.Server) Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.RestPort),
		Handler: h,
	}
	srv.SetKeepAlivesEnabled(true)

	hcsrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HealthzPort),
		Handler: createHealthCheckServiceMux(cfg.HealthzPath),
	}
	hcsrv.SetKeepAlivesEnabled(true)

	gwebsrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.GrpcWebPort),
		Handler: grpcweb.WrapServer(g),
	}
	gwebsrv.SetKeepAlivesEnabled(true)

	dur, err := time.ParseDuration(cfg.ShutdownDuration)
	if err != nil {
		dur = time.Second * 5
	}

	pwt, err := time.ParseDuration(cfg.ProbeWaitTime)
	if err != nil {
		pwt = time.Second * 3
	}

	return &server{
		srv:     srv,
		hcsrv:   hcsrv,
		gwebsrv: gwebsrv,
		grpcsrv: g,
		cfg:     cfg,
		pwt:     pwt,
		sddur:   dur,
	}
}

// ListenAndServe returns a error channel, which includes error returned from api server
// This function start both health check and api server, and the server will close whenever the context receive a Done signal.
// Whenever the server closed, the api server will shutdown after a defined duration (cfg.ProbeWaitTime), while the health check server will shutdown immediately
func (s *server) ListenAndServe(ctx context.Context) chan []error {
	echan := make(chan []error, 1)
	go func() {

		// error channels to keep track server status
		var sech, gech, gwech, hech <-chan error
		// error channels to keep track server status
		var srunning, grunning, gwrunning, hrunning bool

		// start server and define error channels to keep track server status
		if s.srv != nil {
			sech = s.listenAndServe(s.listenAndServeRestAPI)
			srunning = true
		}

		if s.grpcsrv != nil {
			gech = s.listenAndServe(s.listenAndServeGrpcAPI)
			grunning = true
		}

		if s.gwebsrv != nil {
			gwech = s.listenAndServe(s.listenAndServeGrpcWebAPI)
			gwrunning = true
		}

		if s.hcsrv != nil {
			hech = s.listenAndServe(s.hcsrv.ListenAndServe)
			hrunning = true
		}

		time.Sleep(time.Second)

		appendErr := func(errs []error, err error) []error {
			if err != nil {
				return append(errs, err)
			}
			return errs
		}

		errs := make([]error, 0, 3)
		var err error

		for {
			select {
			case <-ctx.Done():
				if hrunning {
					errs = appendErr(errs, s.hcShutdown(ctx))
				}

				if srunning {
					errs = appendErr(errs, s.restShutdown(ctx))
				}

				if grunning {
					errs = appendErr(errs, s.grpcShutdown(ctx))
				}

				if gwrunning {
					errs = appendErr(errs, s.grpcWebShutdown(ctx))
				}

				echan <- appendErr(errs, ctx.Err())
				return

			case err = <-sech:
				if err != nil {
					errs = appendErr(errs, err)
				}
				if hrunning {
					errs = appendErr(errs, s.hcShutdown(ctx))
				}

				if grunning {
					errs = appendErr(errs, s.grpcShutdown(ctx))
				}

				if gwrunning {
					errs = appendErr(errs, s.grpcWebShutdown(ctx))
				}

				echan <- errs
				return
			case err = <-hech:
				if err != nil {
					errs = append(errs, err)
				}

				if srunning {
					errs = appendErr(errs, s.restShutdown(ctx))
				}

				if grunning {
					errs = appendErr(errs, s.grpcShutdown(ctx))
				}

				if gwrunning {
					errs = appendErr(errs, s.grpcWebShutdown(ctx))
				}

				echan <- errs
				return
			case err = <-gech:
				if err != nil {
					errs = append(errs, err)
				}
				if hrunning {
					errs = appendErr(errs, s.hcShutdown(ctx))
				}

				if srunning {
					errs = appendErr(errs, s.restShutdown(ctx))
				}

				if gwrunning {
					errs = appendErr(errs, s.grpcWebShutdown(ctx))
				}

				echan <- errs
				return
			case err = <-gwech:
				if err != nil {
					errs = append(errs, err)
				}
				if hrunning {
					errs = appendErr(errs, s.hcShutdown(ctx))
				}

				if srunning {
					errs = appendErr(errs, s.restShutdown(ctx))
				}

				if grunning {
					errs = appendErr(errs, s.grpcShutdown(ctx))
				}

				echan <- errs
				return
			}
		}
	}()

	return echan
}

// hcShutdown returns error if health check server shutdown unsuccessful
func (s *server) hcShutdown(ctx context.Context) error {
	hctx, hcancel := context.WithTimeout(ctx, s.sddur)
	defer hcancel()
	s.hcsrv.SetKeepAlivesEnabled(false)
	return s.hcsrv.Shutdown(hctx)
}

// restShutdown returns error if rest api server shutdown unsuccessful
func (s *server) restShutdown(ctx context.Context) error {
	time.Sleep(s.pwt)
	sctx, scancel := context.WithTimeout(ctx, s.sddur)
	defer scancel()
	s.srv.SetKeepAlivesEnabled(false)
	return s.srv.Shutdown(sctx)
}

// grpcShutdown returns error if grpc api server shutdown unsuccessful
func (s *server) grpcShutdown(ctx context.Context) error {
	time.Sleep(s.pwt)
	s.grpcsrv.GracefulStop()
	return nil
}

// grpcWebShutdown returns error if grpc web api server shutdown unsuccessful
func (s *server) grpcWebShutdown(ctx context.Context) error {
	time.Sleep(s.pwt)
	sctx, scancel := context.WithTimeout(ctx, s.sddur)
	defer scancel()
	s.gwebsrv.SetKeepAlivesEnabled(false)
	return s.gwebsrv.Shutdown(sctx)
}

// createHealthCheckServiceMux return a *http.ServeMux object
// The function will register the health check server handler for given pattern, and return
func createHealthCheckServiceMux(pattern string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, handleHealthCheckRequest)
	return mux
}

// handleHealthCheckRequest is a handler function for and health check request, which always a HTTP Status OK (200) result
func handleHealthCheckRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		w.Header().Set(ContentType, fmt.Sprintf("%s;%s", TextPlain, CharsetUTF8))
		_, err := fmt.Fprint(w, http.StatusText(http.StatusOK))
		if err != nil {
			glg.Fatal(err)
		}
	}
}

func (s *server) listenAndServe(starter func() error) <-chan error {
	ech := make(chan error, 1)
	go func() {
		ech <- starter()
		close(ech)
	}()
	return ech
}

// listenAndServeGrpcAPI return any error occurred when start a HTTPS server, including any error when loading TLS certificate
func (s *server) listenAndServeGrpcAPI() error {
	cfg, err := NewTLSConfig(s.cfg.TLS)
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.GrpcPort))
	if err == nil && cfg != nil {
		l = tls.NewListener(l, cfg)
	}
	if err != nil {
		glg.Error(err)
	}
	return s.grpcsrv.Serve(l)
}

// listenAndServeGrpcWebAPI return any error occurred when start a HTTPS server, including any error when loading TLS certificate
func (s *server) listenAndServeGrpcWebAPI() error {
	cfg, err := NewTLSConfig(s.cfg.TLS)
	if err == nil && cfg != nil {
		s.gwebsrv.TLSConfig = cfg
	}
	if err != nil {
		glg.Error(err)
	}
	return s.gwebsrv.ListenAndServeTLS("", "")
}

// listenAndServeRestAPI return any error occurred when start a HTTPS server, including any error when loading TLS certificate
func (s *server) listenAndServeRestAPI() error {
	cfg, err := NewTLSConfig(s.cfg.TLS)
	if err == nil && cfg != nil {
		s.srv.TLSConfig = cfg
	}
	if err != nil {
		glg.Error(err)
	}
	return s.srv.ListenAndServeTLS("", "")
}
