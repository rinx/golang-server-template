package router

import (
	"net/http"

	"github.com/kpango/golang-server-template/handler/rest/handler"
)

type Route struct {
	Name        string
	Methods     []string
	Pattern     string
	HandlerFunc handler.Func
}

func NewRoutes(h handler.Handler) []Route {
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
