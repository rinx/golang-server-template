package usecase

import (
	"context"

	"github.com/kpango/golang-server-template/config"
	"github.com/kpango/golang-server-template/handler/grpc"
	"github.com/kpango/golang-server-template/handler/rest"
	"github.com/kpango/golang-server-template/router"
	"github.com/kpango/golang-server-template/service"
)

type Runner interface {
	Start(ctx context.Context) chan error
	Stop(ctx context.Context) error
}

type run struct {
	cfg    config.Config
	server service.Server
}

func New(cfg config.Config) (Runner, error) {
	return &run{
		cfg: cfg,
		server: service.NewServer(cfg.Server,
			grpc.New(),
			router.New(cfg.Server,
				rest.New())),
	}, nil
}

func (t *run) Start(ctx context.Context) chan error {
	return t.server.ListenAndServe(ctx)
}

func (t *run) Stop(ctx context.Context) error {
	return t.server.Shutdown(ctx)
}
