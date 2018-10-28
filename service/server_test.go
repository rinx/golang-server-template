package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/kpango/golang-server-template/config"
	"google.golang.org/grpc"
)

func TestNewServer(t *testing.T) {
	type args struct {
		cfg config.Server
		h   http.Handler
		g   *grpc.Server
	}
	tests := []struct {
		name      string
		args      args
		want      Server
		checkFunc func(got, want Server) error
	}{
		{
			name: "Check health address",
			args: args{
				cfg: config.Server{
					HealthzPath: "/healthz",
					HealthzPort: 8080,
				},
				h: func() http.Handler {
					return nil
				}(),
			},
			want: &server{
				hcsrv: &http.Server{
					Addr: fmt.Sprintf(":%d", 8080),
				},
			},
			checkFunc: func(got, want Server) error {
				if got.(*server).hcsrv.Addr != want.(*server).hcsrv.Addr {
					return fmt.Errorf("Healthz Addr not equals\tgot: %s\twant: %s", got.(*server).hcsrv.Addr, want.(*server).hcsrv.Addr)
				}
				return nil
			},
		},
		{
			name: "Check server address",
			args: args{
				cfg: config.Server{
					GrpcPort:    8083,
					GrpcWebPort: 8082,
					RestPort:    8081,
					HealthzPath: "/healthz",
					HealthzPort: 8080,
				},
				h: func() http.Handler {
					return nil
				}(),
			},
			want: &server{
				srv: &http.Server{
					Addr: fmt.Sprintf(":%d", 8081),
				},
			},
			checkFunc: func(got, want Server) error {
				if got.(*server).srv.Addr != want.(*server).srv.Addr {
					return fmt.Errorf("Server Addr not equals\tgot: %s\twant: %s", got.(*server).srv.Addr, want.(*server).srv.Addr)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewServer(tt.args.cfg, tt.args.h, tt.args.g)
			if err := tt.checkFunc(got, tt.want); err != nil {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_ListenAndServe(t *testing.T) {
	type fields struct {
		srv   *http.Server
		hcsrv *http.Server
		cfg   config.Server
	}
	type args struct {
		ctx        context.Context
		cancelFunc context.CancelFunc
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		checkFunc func(*server, args, chan []error, error) error
		want      error
	}{
		{
			name: "Test servers can start",
			fields: fields{
				srv: &http.Server{
					Handler: func() http.Handler {
						return nil
					}(),
				},
				hcsrv: &http.Server{
					Handler: func() http.Handler {
						return nil
					}(),
				},
			},
			args: func() args {
				ctx, cancelFunc := context.WithCancel(context.Background())
				return args{
					ctx:        ctx,
					cancelFunc: cancelFunc,
				}
			}(),
			checkFunc: func(s *server, args args, got chan []error, want error) error {
				args.cancelFunc()

				if errs := <-got; errs != nil {
					for _, e := range errs {
						if e != want {
							return e
						}
					}
				}
				return nil
			},
			want: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				srv:   tt.fields.srv,
				hcsrv: tt.fields.hcsrv,
				cfg:   tt.fields.cfg,
			}

			e := s.ListenAndServe(tt.args.ctx)
			if err := tt.checkFunc(s, tt.args, e, tt.want); err != nil {
				t.Errorf("server.listenAndServe() Error = %v", err)
			}
		})
	}
}

func Test_server_createHealthCheckServiceMux(t *testing.T) {
	type args struct {
		pattern string
	}
	type test struct {
		name       string
		args       args
		beforeFunc func() error
		checkFunc  func(*http.ServeMux) error
		afterFunc  func() error
		want       http.ServeMux
		wantErr    error
	}
	tests := []test{
		func() test {
			return test{
				name: "Test create server mux",
				args: args{
					pattern: ":8080",
				},
				checkFunc: func(got *http.ServeMux) error {
					if got == nil {
						return fmt.Errorf("serveMux is empty")
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.afterFunc != nil {
				defer func() {
					if err := tt.afterFunc(); err != nil {
						t.Errorf("%v", err)
						return
					}
				}()
			}

			if tt.beforeFunc != nil {
				err := tt.beforeFunc()
				if err != nil {
					t.Errorf("beforeFunc error, error: %v", err)
					return
				}
			}

			got := createHealthCheckServiceMux(tt.args.pattern)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("server.listenAndServeAPI() Error = %v", err)
			}
		})
	}
}

func Test_server_handleHealthCheckRequest(t *testing.T) {
	type args struct {
		rw http.ResponseWriter
		r  *http.Request
	}
	type test struct {
		name       string
		args       args
		beforeFunc func() error
		checkFunc  func() error
		afterFunc  func() error
		wantErr    error
	}
	tests := []test{
		func() test {
			rw := httptest.NewRecorder()

			return test{
				name: "Test handle HTTP GET request health check request",
				args: args{
					rw: rw,
					r:  httptest.NewRequest(http.MethodGet, "/", nil),
				},
				checkFunc: func() error {
					result := rw.Result()
					if header := result.StatusCode; header != http.StatusOK {
						return fmt.Errorf("Header is not correct, got: %v", header)
					}
					if contentType := rw.Header().Get("Content-Type"); contentType != "text/plain;charset=UTF-8" {
						return fmt.Errorf("Content type is not correct, got: %v", contentType)
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.afterFunc != nil {
				defer func() {
					if err := tt.afterFunc(); err != nil {
						t.Errorf("%v", err)
						return
					}
				}()
			}

			if tt.beforeFunc != nil {
				err := tt.beforeFunc()
				if err != nil {
					t.Errorf("beforeFunc error, error: %v", err)
					return
				}
			}

			handleHealthCheckRequest(tt.args.rw, tt.args.r)
			if err := tt.checkFunc(); err != nil {
				t.Errorf("error: %v", err)
			}
		})
	}
}

func Test_server_listenAndServeAPI(t *testing.T) {
	type fields struct {
		srv   *http.Server
		hcsrv *http.Server
		cfg   config.Server
	}
	type test struct {
		name       string
		fields     fields
		beforeFunc func() error
		checkFunc  func(*server, error) error
		afterFunc  func() error
		want       error
	}
	tests := []test{
		func() test {
			keyKey := "dummy_key"
			key := "./assets/dummyServer.key"
			certKey := "dummy_cert"
			cert := "./assets/dummyServer.crt"

			return test{
				name: "Test server startup",
				fields: fields{
					srv: &http.Server{
						Handler: func() http.Handler {
							return nil
						}(),
						Addr: fmt.Sprintf(":%d", 9999),
					},
					cfg: config.Server{
						GrpcPort:    9997,
						GrpcWebPort: 9998,
						RestPort:    9999,
						TLS: config.TLS{
							Enabled: true,
							CertKey: certKey,
							KeyKey:  keyKey,
						},
					},
				},
				beforeFunc: func() error {
					err := os.Setenv(keyKey, key)
					if err != nil {
						return err
					}
					err = os.Setenv(certKey, cert)
					if err != nil {
						return err
					}
					return nil
				},
				checkFunc: func(s *server, want error) error {
					// listenAndServeAPI function is blocking, so we need to set timer to shutdown the process
					go func() {
						time.Sleep(time.Second * 3)
						s.srv.Shutdown(context.Background())
					}()

					got := s.listenAndServeRestAPI()

					if got != want {
						return fmt.Errorf("got:\t%v\nwant:\t%v", got, want)
					}
					return nil
				},
				afterFunc: func() error {
					os.Remove(keyKey)
					os.Remove(certKey)
					return nil
				},
				want: http.ErrServerClosed,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.afterFunc != nil {
				defer func() {
					if err := tt.afterFunc(); err != nil {
						t.Errorf("%v", err)
						return
					}
				}()
			}

			if tt.beforeFunc != nil {
				err := tt.beforeFunc()
				if err != nil {
					t.Errorf("beforeFunc error, error: %v", err)
					return
				}
			}

			if err := tt.checkFunc(&server{
				srv:   tt.fields.srv,
				hcsrv: tt.fields.hcsrv,
				cfg:   tt.fields.cfg,
			}, tt.want); err != nil {
				t.Errorf("server.listenAndServeAPI() Error = %v", err)
			}
		})
	}
}
