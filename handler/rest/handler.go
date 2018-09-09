package rest

import (
	"net/http"
)

type Handler interface {
	Sample(http.ResponseWriter, *http.Request) error
}

type Func func(http.ResponseWriter, *http.Request) error

type handler struct {
}

func New() Handler {
	return nil
}

func (h *handler) Sample(http.ResponseWriter, *http.Request) error {
	return nil
}
