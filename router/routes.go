package router

import (
	"net/http"

	"github.com/kpango/golang-server-template/handler/rest"
)

type Route struct {
	Name        string
	Methods     []string
	Pattern     string
	HandlerFunc rest.Func
}

func NewRoutes(h rest.Handler) []Route {
	return []Route{
		{
			"Sample Handler",
			[]string{
				http.MethodGet,
			},
			"/sample",
			h.Sample,
		},
	}
}
